What's this?

Near realtime replication of Google Secret Manager (GSM) secrets into
kubernetes secrets.

Event notifications are enabled in GSM. This worload runs in your cluster and
listens for the notifications on a Google Pub/Sub subscription.

_Well... when it's done_

## Setup

See https://cloud.google.com/secret-manager/docs/event-notifications for
extra details on setting up event notifications. I detail it briefly here

If we name the topic and subscription `secets.events` and `secrets.events.gsm-pubsub`
respectively, the setup can look a bit like

```shell
export PROJECT_ID=your-project-id

gcloud beta services identity create \
    --service "secretmanager.googleapis.com" \
    --project "${PROJECT_ID}"

# ^ This will print out a service account which can be constructed as follows ˅

PROJECT_NUMBER=gcloud projects describe $PROJECT_ID --format='value(projectNumber)'
export SM_SERVICE_ACCOUNT=service-${PROJECT_NUMBER}@gcp-sa-secretmanager.iam.gserviceaccount.com


gcloud pubsub topics create "projects/${PROJECT_ID}/topics/secrets.events"

gcloud pubsub topics add-iam-policy-binding secrets.events \
    --member "serviceAccount:${SM_SERVICE_ACCOUNT}" \
    --role "roles/pubsub.publisher"

gcloud pubsub subscriptions create \
    "projects/${PROJECT_ID}/subscriptions/secrets.events.gsm-pubsub" \
    --topic "projects/${PROJECT_ID}/topics/secrets.events"
```

Then when creating new secrets the above topic should be specified
```shell
gcloud secrets create test-secret \
    --topics "projects/${PROJECT_ID}/topics/secrets.events"
```


This application also needs some permissions.
I did my testing in a [kind](https://kind.sigs.k8s.io/) cluster, so here I've
called the service account `kind-cluster-sm`, but adjust accordingly.

```shell
gcloud iam service-accounts create kind-cluster-sm --project $PROJECT_ID

# Allow reading the subscription
gcloud pubsub subscriptions add-iam-policy-binding secrets.events.gsm-pubsub \
    --member "serviceAccount:kind-cluster-sm@$PROJECT_ID.iam.gserviceaccount.com" \
    --role "roles/pubsub.subscriber"

# Allow accessing scerets
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --role roles/secretmanager.secretAccessor \
  --member "serviceAccount:kind-cluster-sm@$PROJECT_ID.iam.gserviceaccount.com" \
  --project $PROJECT_ID
```

To deploy outside of GKE, e.g. in a namespace called `gsm`
```shell
mkdir -p overlays/local && cd overlays/local

gcloud iam service-accounts keys create key.json \
  --iam-account kind-cluster-sm@$PROJECT_ID.iam.gserviceaccount.com

cat > kustomization.yaml <<EOF
namespace: gsm
resources:
- github.com/cakemanny/gsm-pubsub/bases/main
- namespace.yaml
patches:
- deployment.yaml
secretGenerator:
- name: gsm-application-credentials
  files:
  - key.json
EOF

cat > namespace.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: gsm
EOF

cat > deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gsm-pubsub
spec:
  template:
    spec:
      containers:
      - name: gsm-pubsub
        env:
        - name: PROJECT_ID
          value: $PROJECT_ID
        - name: SUBSCRIPTION
          value: secrets.events.gsm-pubsub
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: /var/secrets/key.json
        volumeMounts:
          - mountPath: /var/secrets/
            name: gcp-key
            readOnly: true
      volumes:
      - name: gcp-key
        secret:
          secretName: gsm-application-credentials
          optional: false
EOF

kubectl apply -k .
```

Instead if deploying in GKE with workload identity, it might look a bit
more like the below. I keep the service account name `kind-cluster-sa`
for ease of matching with further above.

```shell
cat > kustomization.yaml <<EOF
namespace: gsm
resources:
- github.com/cakemanny/gsm-pubsub/bases/main
- namespace.yaml
patches:
- deployment.yaml
- service-account.yaml
EOF

cat > namespace.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: gsm
EOF

cat > service-account.yaml <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gsm-sa
  annotations:
    iam.gke.io/gcp-service-account: kind-cluster-sm@PROJECT_ID.iam.gserviceaccount.com
EOF

cat > deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gsm-pubsub
spec:
  template:
    spec:
      containers:
      - name: gsm-pubsub
        env:
        - name: PROJECT_ID
          value: $PROJECT_ID
        - name: SUBSCRIPTION
          value: secrets.events.gsm-pubsub
EOF

kubectl apply -k .
```
