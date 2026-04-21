# Bank Integration Status — Venezuelan Banking Ecosystem

> **Last Updated:** April 2026
> **Project Name:** TBD (formerly "Faloppa Payments" — placeholder)

> [!IMPORTANT]
> Only **BNC** has confirmed, verified API documentation. All other entries are based on public web research and will require direct contact with each bank's corporate/developer team to obtain actual API specs.

---

## 🟢 TIER 1 — Confirmed Documentation

### BNC — Banco Nacional de Crédito (Code: 0191)
| Item | Value |
|---|---|
| **Documentation** | ✅ Complete and verified |
| **Sources** | ESolutions API v4.1 (Postman), NotificationPush V2.0 (Postman) |
| **C2P Endpoint** | `POST /api/MobPayment/SendC2P` |
| **Auth** | Daily WorkingKey via Logon (MasterKey for auth) |
| **Encryption** | Custom AES/PBKDF2 body encryption (UTF-16LE, fixed salt) |
| **Notifications** | ✅ Push webhooks (NotificationPush V2.0) |
| **OTP** | Generated via BNCNET app/portal; valid until 11:59 PM |
| **Multi-tenant** | ChildClientID + BranchID hierarchy |
| **Note** | BNC absorbed BOD (Banco Occidental de Descuento) in 2022 |
| **Status** | 🚀 **Ready to implement** |

---

## 🟡 TIER 2 — Public Developer Portal Exists (Requires Registration)

### Mercantil Banco Universal (Code: 0105)
| Item | Value |
|---|---|
| **Documentation** | ⚠️ Portal exists, requires registration |
| **Developer Portal** | `https://www.mercantilbanco.com/mercprod/apiportal/` |
| **Product** | "Botón de Pagos Móviles (C2P) y Vuelto" |
| **Auth** | `X-IBM-Client-Id` header + app credentials |
| **Format** | OpenAPI/Swagger downloadable (JSON/YAML) |
| **Endpoint Domain** | `apimbu.mercantilbanco.com` (production) |
| **Encryption** | AES-256 for sensitive fields |
| **Support** | apisupport@bancomercantil.com / (0212) 503.25.99 |
| **Requirements** | Must be Mercantil account holder + "Mercantil en Línea Empresas" |
| **Status** | 📋 **Register on portal to obtain full docs** |

### Banco Plaza (Code: 0138)
| Item | Value |
|---|---|
| **Documentation** | ⚠️ Open Banking portal exists, requires registration |
| **Portal** | `https://www.bancoplaza.com` (Open Banking section) |
| **C2P Endpoints** | `/v0/pagos/c2p/` (charge), `/v0/pagos/c2p/bancos` (bank list), `/v0/pagos/c2p/{rif}/{ref}` (query) |
| **QA Domain** | `https://apiqa.bancoplaza.com:8585` |
| **Prod Domain** | `https://api.bancoplaza.com` |
| **Auth** | HMAC-SHA384 digital signature (`API-KEY`, `API-KEY-SECRET`, `NONCE`, `API-SIGNATURE`) |
| **Channels** | 20=POS, 21=MERCHANT, 23=BOTÓN DE PAGO |
| **Idempotency** | `id-externo` field |
| **Contact** | 0212-2192600, centrodeatencion@bancoplaza.com |
| **Status** | 📋 **Register on portal to obtain full docs** |

---

## 🟠 TIER 3 — APIs Exist But No Public Portal (Contact Required)

### Banesco Banco Universal (Code: 0134)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public portal; requires executive contact |
| **Known Products** | API Banesco Pagos, Vuelto P2P, Confirmación de Transacción |
| **Access** | Via "Desarrolladores" section on banesco.com (gated) |
| **C2P** | Available through "Comercio Afiliado" program |
| **Suite** | Suite BanescoPagos (app for juridical clients) |
| **How to access** | Contact assigned Ejecutivo de Cuenta or Ejecutivo Comercial de Banesco Comercio Electrónico |
| **Status** | 📞 **Must contact bank executive** |

### Bancaribe (Code: 0114)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public portal; managed via corporate channels |
| **Known Services** | `queryPaymentB2P` (payment verification), C2P via API Management |
| **Auth** | OAuth2 for token generation |
| **Protocol** | REST (JSON) and SOAP |
| **Partnerships** | Neerü (Suiche7B), Siteff, Megasoft, Novared |
| **Contact** | mdpagos@bancaribe.com.ve, (0424)1544110, (0212)9546406 |
| **Requirements** | "Mi Conexión Bancaribe Jurídica" affiliation |
| **Status** | 📞 **Must contact bank — has API capabilities** |

### Banplus (Code: 0174)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public portal |
| **Known APIs** | Validación pago móvil, Validación transferencias, Notificación transacciones, Envío de pago (vuelto), Consulta saldos |
| **Portal Section** | "Integraciones APIs" on banplus.com |
| **Requirements** | "Pago Plus Comercio" affiliation via "Banplus On Line Empresas" |
| **Contact** | Contact Gerente de Negocios at Banplus |
| **Status** | 📞 **Must contact bank** |

### BFC — Banco Fondo Común (Code: 0151)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public API |
| **Known Products** | Botón de Pago C2P, Enlace de Pago C2P, POS Virtual C2P |
| **Apps** | "BFC Comercio" (Android/iOS), Web Admin C2P |
| **Requirements** | "+BFC en Línea – Empresas" affiliation |
| **Contact** | SoporteInternetBanking@bfc.com.ve |
| **Status** | 📞 **Must contact bank** |

---

## 🔴 TIER 4 — No Public API / Executive-Only Access

### Banco de Venezuela (Code: 0102)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public developer portal |
| **Known Products** | "Botón de pagoBDV", "PagomóvilBDV Comercio" |
| **Innovation Hub** | hub_bdv_innova@banvenez.com (for startups) |
| **Access** | Executive contact only; requires juridical account |
| **Contact** | atencion_clientejuridico@banvenez.com |
| **Status** | 🚫 **No public API — executive contact only** |

### BBVA Provincial (Code: 0108)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public developer portal for VZ C2P |
| **Known Products** | "Dinero Rápido" C2P (app-only), Provinet Empresas |
| **Global Portal** | BBVA API Market (bbvaapimarket.com) — international, not VZ C2P |
| **Access** | Through Provinet Empresas or assigned executive |
| **Status** | 🚫 **No public API — executive contact only** |

### Banco Exterior (Code: 0115)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public API |
| **Known Products** | "Exterior NEXO Pago Móvil a Comercios (P2C)", "Cobro Comercio (C2P)" |
| **Requirements** | "Exterior NEXO Jurídico" affiliation |
| **Contact** | Soporte Nexo Jurídicos (0212) 501-5500 |
| **Status** | 🚫 **No public API — executive contact only** |

### Bancamiga (Code: 0172)
| Item | Value |
|---|---|
| **Documentation** | ❌ No public API documentation |
| **Known Products** | "AdminPagos" (C2P management), "BotonPago" (e-commerce) |
| **Contact** | 0500-TUBANCA (0500-8822622), atencion.alcliente@bancamiga.com |
| **Status** | 🚫 **No public API — requires formal process with bank** |

### Banco Bicentenario (Code: 0175)
| Item | Value |
|---|---|
| **Documentation** | ❌ Nothing found publicly |
| **Status** | 🚫 **No information available** |

---

## 🔵 Smaller Banks (Not Investigated in Detail)

| Bank | Code | Notes |
|---|---|---|
| Venezolano de Crédito | 0104 | No public API found |
| Sofitasa | 0137 | No public API found |
| 100% Banco | 0156 | No public API found |
| Banco Activo | 0171 | No public API found |
| Mi Banco | 0169 | No public API found |
| Banco Caroní | 0128 | No public API found |
| BOD | 0116 | ❌ **Absorbed by BNC in 2022 — no longer exists** |

---

## 🔗 Payment Network Operators

| Network | Role | Notes |
|---|---|---|
| **Suiche 7B** | Interbank switch | Routes interbank Pago Móvil transactions |
| **Conexus** | Interbank switch | Alternative interbank routing network |
| **NovaRed** | Interbank switch | Third interbank routing option |
| **Megasoft (Merchant)** | PSP/Aggregator | Pre-built C2P button integration, may simplify multi-bank |

> [!TIP]
> Integration is done **bank-by-bank**, not through the switch operators directly. However, some PSPs like Megasoft already aggregate multiple banks.

---

## MVP Strategy Recommendation

Given the documentation landscape:

1. **Start with BNC (0191)** — Only bank with confirmed, complete documentation ✅
2. **Register on Mercantil portal (0105)** — Download OpenAPI specs, implement second adapter
3. **Register on Banco Plaza portal (0138)** — Surprising find: has Open Banking with documented C2P endpoints
4. **Contact Banesco (0134)** via executive — Known APIs exist, just gated
5. **Contact Bancaribe (0114)** — Has OAuth2 APIs, documented internally

**Priority order for implementation:**
```
BNC (ready) → Mercantil (register) → Banco Plaza (register) → Banesco (contact) → Bancaribe (contact)
```
