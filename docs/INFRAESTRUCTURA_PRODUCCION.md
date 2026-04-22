# Infraestructura de Producción — Faloppa Payment Gateway

> **Última actualización:** 2026-04-21  
> **Estado:** Pre-producción — Esperando credenciales bancarias  
> **Decisión pendiente:** Aprobar presupuesto de infraestructura

---

## Contexto

Este sistema procesa pagos C2P (Cobro a Persona) con bancos venezolanos.  
Maneja credenciales bancarias de terceros (comercios) y coordina transacciones monetarias reales.  
La infraestructura debe proteger dinero ajeno y cumplir con estándares de certificación BNC y regulación SUDEBAN.

---

## Qué servicios necesitamos y por qué

### 1. Servidor de aplicación — GCP Cloud Run (~$15-20/mes)

**Qué es:** Donde corre el código Go del Gateway.  
**Por qué Cloud Run y no un VPS tradicional:**
- Escala automáticamente si hay picos de tráfico (ej. temporada de pagos masivos)
- Escala a cero si no hay tráfico (no pagas por servidores ociosos)
- Google gestiona los parches de seguridad del sistema operativo
- Despliegue inmutable: la misma imagen Docker probada en QA es la que va a producción

**¿Se puede recortar?** No. El código tiene que correr en algún lado. Cloud Run es la opción más barata y segura para Go.

---

### 2. Base de datos — GCP Cloud SQL PostgreSQL (~$55-150/mes)

**Qué es:** Donde se guardan las transacciones, comercios y credenciales cifradas.  
**Por qué Cloud SQL y no Supabase Pro:**
- Cloud SQL te da **IP privada** — la base de datos NO está expuesta a internet
- Solo el Gateway (dentro de la misma red privada de Google) puede conectarse
- Si alguien roba tu contraseña de la DB, no puede conectarse porque no tiene acceso a la red
- Supabase Pro ($25/mes) NO ofrece red privada. Tu DB queda expuesta al internet público con solo una contraseña protegiéndola. Para red privada en Supabase necesitas el plan Teams ($599/mes)

**Dos niveles de inversión:**

| Nivel | Costo | Qué obtienes | Riesgo |
|---|---|---|---|
| **Sin HA** | ~$55/mes | 1 servidor de DB en red privada, backups automáticos diarios | Si el data center falla, hay ~5-15 min de downtime mientras GCP restaura |
| **Con HA (Multi-AZ)** | ~$150/mes | 2 servidores sincronizados en data centers distintos, failover automático en segundos | Prácticamente cero downtime |

**¿Se puede recortar?** Puedes empezar sin HA ($55/mes) e ir a HA cuando el volumen lo justifique. Lo que NO puedes recortar es la IP privada — es la diferencia entre "seguro" y "expuesto".

---

### 3. IP estática de salida — GCP Cloud NAT (~$35/mes)

**Qué es:** Un componente de red que le da a tu servidor una IP fija para comunicarse con BNC.  
**Por qué es obligatorio:**
- BNC exige saber tu IP para meterla en su lista blanca (Whitelist)
- Sin esto, BNC rechaza tus conexiones
- También garantiza que tu servidor no tiene IP pública directa (está oculto detrás del NAT)

**¿Se puede recortar?** No. BNC lo exige. Sin IP fija no hay certificación.

---

### 4. Criptografía de hardware — GCP Cloud KMS + Secret Manager (~$2/mes)

**Qué es:** Un chip de hardware (HSM) dentro de Google que cifra y descifra las credenciales bancarias de tus comercios.  
**Por qué es obligatorio:**
- Las credenciales bancarias (MasterKey, ClientGUID) que tus comercios te confían se cifran con una clave que NUNCA sale del chip de hardware
- Ni tú, ni un hacker que comprometa tu servidor, ni un empleado desleal puede extraer la clave maestra
- El código Go ya tiene la interfaz `KMSClient` preparada — solo hay que apuntar al servicio real en vez del mock

**¿Se puede recortar?** No. Actualmente usamos un `MockKMS` que cifra con una clave en memoria. En producción con dinero real, esto es inaceptable. Si alguien compromete el servidor, tiene acceso a todas las credenciales bancarias de todos tus comercios.

---

### 5. Protección perimetral — Cloudflare Pro (~$20/mes)

**Qué es:** Un escudo que se pone DELANTE de tu servidor y filtra el tráfico malicioso.  
**Qué incluye:**
- **WAF (Web Application Firewall):** Bloquea ataques automáticos (inyecciones SQL, XSS, bots)
- **DDoS Protection:** Si alguien intenta tumbar tu servidor con millones de peticiones, Cloudflare las absorbe
- **SSL/TLS:** Certificados HTTPS automáticos y gratuitos
- **Reglas de IP:** Puedes configurar que el endpoint `/v1/webhooks/bnc` SOLO acepte tráfico de las IPs oficiales de BNC

**¿Se puede recortar?** El WAF sí es recortable si aceptas el riesgo. Cloudflare Free te da DNS + SSL + DDoS básico. Pero pierdes las reglas de firewall avanzadas y la protección contra inyecciones. Para un sistema que maneja dinero, los $20/mes son una póliza de seguro barata.

---

### 6. Dominio — Cloudflare Registrar (~$10-15/año)

**Qué es:** Tu dirección en internet (ej. `api.faloppa.com`).  
**Configuración:**
- `dev-api.tudominio.com` → Ambiente de desarrollo (para certificación con BNC)
- `api.tudominio.com` → Ambiente de producción

**¿Se puede recortar?** No. BNC necesita URLs reales para enviar webhooks.

---

## Presupuesto consolidado

### Opción recomendada para arrancar: ~$130/mes

| # | Servicio | Proveedor | Costo/mes | ¿Obligatorio? |
|---|---|---|---|---|
| 1 | Cloud Run (Go Gateway) | GCP | $15 | ✅ Sí |
| 2 | Cloud SQL PostgreSQL (sin HA, IP privada) | GCP | $55 | ✅ Sí |
| 3 | Cloud NAT (IP estática) | GCP | $35 | ✅ Sí |
| 4 | Cloud KMS + Secret Manager | GCP | $2 | ✅ Sí |
| 5 | Cloudflare Pro (WAF + DNS) | Cloudflare | $20 | Recomendado |
| 6 | Dominio | Cloudflare | $1 | ✅ Sí |
| | **Total** | | **~$128** | |

### Upgrade futuro: ~$265/mes

Cuando el volumen de transacciones justifique alta disponibilidad:

| Upgrade | Costo adicional |
|---|---|
| Cloud SQL → HA Multi-AZ | +$95/mes |
| Memorystore Redis (idempotencia persistente) | +$35/mes |

---

## Dónde comprar (solo 2 proveedores)

1. **Google Cloud Platform** — [cloud.google.com](https://cloud.google.com)
   - Todo lo de infraestructura: servidores, base de datos, red, criptografía
   - Te regalan $300 en créditos los primeros 3 meses
   
2. **Cloudflare** — [cloudflare.com](https://www.cloudflare.com)
   - Dominio + DNS + WAF + SSL
   - Compra el dominio directamente aquí (precio al costo, sin markup)

---

## Qué NO necesitas comprar

- ❌ **Hostinger / GoDaddy / Namecheap** — Solo para hosting compartido, no para bancos
- ❌ **Heroku** — No tiene red privada ni IPs estáticas
- ❌ **Supabase Pro ($25)** — No te da red privada. Para eso Cloud SQL es mejor y más barato que Supabase Teams ($599)
- ❌ **AWS** — Podrías, pero mezclar proveedores de nube aumenta la complejidad sin beneficio. Mejor todo en GCP
- ❌ **Datadog / New Relic** — GCP Cloud Logging es suficiente para empezar ($0-5/mes incluido en GCP)
