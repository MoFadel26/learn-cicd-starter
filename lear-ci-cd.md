Workload Identity Federation
We want to add an Artifact Registry step to the continuous delivery pipeline. gcloud builds submit works from your laptop using your Google account identity. Which identity would you use from GitHub Actions?

We can use Workload Identity Federation (WIF). Federation here means allowing an identity from one system to be trusted by another. We want Google Cloud to trust an identity from GitHub Actions. This is much better than long-lived credentials that need to be kept secret and regularly rotated. These two steps have proven challenging for many an organization.

How It Works
The federation between GitHub and GCP is powered by OpenID Connect (OIDC).

In the GitHub Actions workflow, GitHub makes a signed OIDC token available. The token contains information about the running workflow, including its repository and branch.

The google-github-actions/auth GitHub Action sends the OIDC token and the service account we want to use to GCP.

If the GitHub identity is allowed to impersonate the service account, GCP returns short-lived credentials for the service account.

Now the workflow can use gcloud builds submit with the permissions of the service account.

Pools and Providers
We configure two things in GCP:

A workload identity pool groups external identities.

A workload identity provider defines how to verify a token from an external identity provider, like GitHub.

We don't want to trust EVERY token from GitHub - that would allow access from ANY workflow from ANY GitHub user. You can write rules in the provider to restrict access to your repository or a particular branch.

Assignment
There are going to be a lot of gcloud commands here. Making things secure does take extra steps.

Creating a Service Account
Go to the IAM & Admin Service Accounts section of the GCP console.

Create a service account and name it "Cloud Run Deployer" with these permissions:

Cloud Build Editor

Cloud Build Service Account

Cloud Run Admin

Service Account User

Viewer

Configuring WIF
The Configure Workload Identity Federation with deployment pipelines documentation explains the APIs we're enabling and the resources we're creating.

To start configuring WIF, go to the GCP Enable API wizard.

Confirm your GCP project is selected.

Enable the pre-selected APIs required for WIF:

Identity and Access Management (IAM) API

Cloud Resource Manager API

IAM Service Account Credentials API

Security Token Service API

From your repo's root, create temporary shell variables to use in the gcloud setup commands.

GCP_PROJECT_ID="$(gcloud config get-value project)"
GCP_SERVICE_ACCOUNT_EMAIL="cloud-run-deployer@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
GITHUB_REPO="$(git remote get-url origin | sed -E 's#.*github.com[:/]##; s#\.git$##')"

echo "
GCP Project:         ${GCP_PROJECT_ID}
GCP Service Account: ${GCP_SERVICE_ACCOUNT_EMAIL}
GitHub Repo:         ${GITHUB_REPO}
"

Create a workload identity pool.

gcloud iam workload-identity-pools create "github" \
  --location="global" \
  --display-name="GitHub Actions Pool"

Create a provider for GitHub in the workload identity pool. This maps claims in the GitHub OIDC token to attributes we can refer to in IAM restrictions. The attribute condition used here only accepts tokens from workflows run in your GitHub repo.

The issuer URL identifies GitHub's OIDC provider. If you're curious, you can see the public keys used to verify GitHub's OIDC tokens.

gcloud iam workload-identity-pools providers create-oidc "my-repo" \
  --location="global" \
  --workload-identity-pool="github" \
  --display-name="My GitHub repo Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository == '${GITHUB_REPO}'" \
  --issuer-uri="https://token.actions.githubusercontent.com"

Save the full ID of the workload identity pool in a temporary environment variable.

export WORKLOAD_IDENTITY_POOL="$(
  gcloud iam workload-identity-pools describe "github" \
    --location="global" \
    --format="value(name)"
)"

Allow identities from your GitHub repo in the workload identity pool to impersonate the "Cloud Run Deployer" service account. The member of the IAM policy binding is a principalSet that matches workload identities with a specific attribute value. Here, it matches identities in the new pool whose attribute.repository value is your GitHub repo. This is why we configured an attribute mapping in the workload identity provider.

gcloud iam service-accounts add-iam-policy-binding \
  ${GCP_SERVICE_ACCOUNT_EMAIL} \
  --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/${WORKLOAD_IDENTITY_POOL}/attribute.repository/${GITHUB_REPO}"

Run and submit the CLI tests from the root of your repo.
