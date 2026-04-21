# Deployment Guide — Payment Gateway

## Step 1: GitHub Repository

1. Go to [github.com/new](https://github.com/new)
2. Repository name: `payment-gateway`
3. Visibility: **Private**
4. Do NOT initialize with README (we already have one)
5. Click "Create repository"
6. Run these commands in terminal:

```bash
cd ~/Projects/faloppa-payments
git remote add origin git@github.com:YOUR_USERNAME/payment-gateway.git
git push -u origin main
```

---

## Step 2: Supabase Project

1. Go to [supabase.com/dashboard](https://supabase.com/dashboard)
2. Click "New Project"
3. Name: `payment-gateway`
4. Database Password: (save this securely)
5. Region: Choose closest to your users
6. Click "Create new project"

### Run Migration
1. In Supabase Dashboard → **SQL Editor**
2. Click "New query"
3. Copy the entire contents of `supabase/migrations/001_core_schema.sql`
4. Click **Run**
5. Verify: go to **Table Editor** — you should see `merchants`, `merchant_bank_configs`, `transactions`

### Run Seed (Development Only)
1. In SQL Editor → New query
2. Copy the contents of `supabase/seed.sql`
3. Click **Run**

### Get Database URL
1. Go to **Settings** → **Database**
2. Copy the **Connection string** (URI format)
3. It looks like: `postgres://postgres:PASSWORD@db.XXXXX.supabase.co:5432/postgres`
4. Create your `.env` file:

```bash
cp .env.example .env
# Edit .env and paste the DATABASE_URL
```

---

## Step 3: GCP Project (for Production)

### Create Project
1. Go to [console.cloud.google.com](https://console.cloud.google.com)
2. Create a new project: `payment-gateway`
3. Enable APIs:
   - Cloud Run API
   - Cloud KMS API
   - Artifact Registry API

### Create KMS Key
1. Go to **Security** → **Key Management**
2. Click "Create key ring"
   - Name: `payment-gateway`
   - Location: `us-central1` (or your region)
3. Click "Create key"
   - Name: `merchant-credentials`
   - Purpose: Symmetric encrypt/decrypt
4. Copy the key resource name — it looks like:
   ```
   projects/YOUR_PROJECT/locations/us-central1/keyRings/payment-gateway/cryptoKeys/merchant-credentials
   ```
5. Add this to your `.env` as `KMS_KEY_RESOURCE_NAME`

### Deploy to Cloud Run
1. Install gcloud CLI: `brew install google-cloud-sdk`
2. Authenticate: `gcloud auth login`
3. Set project: `gcloud config set project YOUR_PROJECT_ID`
4. Deploy:
   ```bash
   gcloud run deploy payment-gateway \
     --source . \
     --region us-central1 \
     --set-env-vars "DATABASE_URL=postgres://...,KMS_KEY_RESOURCE_NAME=projects/...,APP_ENV=production"
   ```

### Set Up GitHub Actions Secrets
In your GitHub repo → Settings → Secrets → Actions, add:
- `GCP_PROJECT_ID`
- `GCP_WIF_PROVIDER` (Workload Identity Federation)
- `GCP_SA_EMAIL` (Service Account email)

---

## Step 4: Generate Your First API Key

```bash
go run ./cmd/keygen -name "Your Company Name"
```

Save the API key. Run the SQL INSERT in Supabase SQL Editor.

---

## Step 5: Test Locally

```bash
# Start the server
export DATABASE_URL="postgres://..."
export KMS_KEY_RESOURCE_NAME="mock"
go run ./cmd/gateway

# In another terminal, test the health endpoint
curl http://localhost:8080/health

# Test a charge (will use BNC test mode)
curl -X POST http://localhost:8080/v1/charges/c2p \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Idempotency-Key: test-1" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00,
    "payer": {
      "bank_code": "0191",
      "phone": "04141234567",
      "id_document": "V12345678",
      "otp": "A1B2C3"
    },
    "idempotency_key": "test-1"
  }'
```

---

## Step 6: BNC Credentials

Contact BNC to obtain:
- `ClientGUID` — your commerce identifier
- `MasterKey` — your encryption master key
- `Terminal` — your terminal identifier

These will be encrypted and stored in `merchant_bank_configs` via the Dashboard (Phase 2).
