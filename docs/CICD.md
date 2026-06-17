# CI/CD

This repo uses a split CI/CD pipeline. **CI** (GitHub Actions) builds and signs an
immutable artifact; **CD** is GitOps — a separate config repo is the source of truth
and Argo CD reconciles it onto the clusters. CI never touches a cluster.

## Workflows

| File | Trigger | What it does | Deploys? |
|------|---------|--------------|----------|
| `pr-validation.yaml` | `pull_request` → `main` | lint, vet, test+race, govulncheck, Trivy fs, Semgrep, gitleaks, **build-only** (no push) | no |
| `release.yaml` | `push` → `main` | re-runs CI gate, builds+pushes by digest, Trivy image scan, SBOM + SLSA provenance, Cosign sign, **auto-deploy to staging** | staging |
| `promote.yaml` | manual `workflow_dispatch` | opens an approved PR in the GitOps repo to move the **same digest** to production | production |
| `reusable-ci-checks.yaml` | `workflow_call` | the shared quality+security gate | — |
| `reusable-build-sign.yaml` | `workflow_call` | the shared build / scan / sign logic | — |

Flow: open PR → `pr-validation` must pass + reviews → merge → `release` builds, signs,
deploys staging → run `promote` (approval) → PR merged in GitOps repo → Argo CD rolls prod.

## Required GitHub secrets

Set in **Settings → Secrets and variables → Actions** (repo or org level).

| Secret | Used by | Description |
|--------|---------|-------------|
| `WORKLOAD_IDENTITY_PROVIDER` | release | Full WIF provider resource name (see OIDC below) |
| `GCP_SERVICE_ACCOUNT` | release | Service account email CI impersonates |
| `GCP_REGION` | release | Artifact Registry region, e.g. `us-central1` |
| `GCP_PROJECT_ID` | release | GCP project id |
| `GAR_REPOSITORY` | release | Artifact Registry repo name |
| `GITOPS_REPO` | release, promote | `owner/repo` of the GitOps config repo |
| `GITOPS_PAT` | release, promote | Fine-grained PAT with Contents+PR write on the GitOps repo |
| `GITLEAKS_LICENSE` | ci-checks | Only for **private** repos in a GitHub org (optional) |

No long-lived cloud keys are stored: GCP auth is via OIDC/Workload Identity Federation,
and image signing is keyless (Sigstore).

## OIDC / Workload Identity Federation setup (one time)

```bash
PROJECT_ID=your-project
POOL=github-pool
PROVIDER=github-provider
SA=gha-deployer@${PROJECT_ID}.iam.gserviceaccount.com
REPO=sverma/gtod   # owner/repo

gcloud iam workload-identity-pools create "$POOL" \
  --project="$PROJECT_ID" --location=global --display-name="GitHub Actions"

gcloud iam workload-identity-pools providers create-oidc "$PROVIDER" \
  --project="$PROJECT_ID" --location=global --workload-identity-pool="$POOL" \
  --display-name="GitHub" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='${REPO}'"

# Let the GitHub repo impersonate the deployer service account
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
gcloud iam service-accounts add-iam-policy-binding "$SA" \
  --project="$PROJECT_ID" --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL}/attribute.repository/${REPO}"

# Grant the SA permission to push to Artifact Registry
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${SA}" --role=roles/artifactregistry.writer
```

Then set the secret:
`WORKLOAD_IDENTITY_PROVIDER = projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-pool/providers/github-provider`

## GitHub Environments (approval gates)

Create two environments under **Settings → Environments**:

- `staging` — used by `release.yaml` (optionally add reviewers/wait timer).
- `production` — used by `promote.yaml`. Add **required reviewers**; promotion pauses
  for approval before the PR is even opened in the GitOps repo.

## In-cluster CD (GitOps repo)

- `deploy/policies/verify-images-kyverno.yaml` — Kyverno policy that rejects any
  `worlddatetime` image not signed by this repo's workflows. This is what makes the
  Cosign signature actually enforced at admission time.
- `deploy/rollouts/worlddatetime-rollout.yaml` — Argo Rollouts canary (5%→25%→50%→100%)
  with a Prometheus success-rate analysis gate for automated rollback.

The GitOps repo is expected to contain `apps/worlddatetime/overlays/{stage,prod}/image.yaml`
with a line like `...worlddatetime:<tag>@sha256:<digest>` that the pipelines rewrite.
