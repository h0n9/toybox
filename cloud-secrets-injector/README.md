# Cloud Secrets Injector

This simple yet powerful tool named **Cloud Secrets Injector** aims to simplify
the way to inject secrets strored on Cloud-based secrets managers into
Kubernetes Pods, functioning as [HashiCorp Vault's Agent Sidecar
Injector](https://www.vaultproject.io/docs/platform/k8s/injector).

## Supported Cloud Providers
- AWS(Amazon Web Services): Secrets Manager

## Usage

### Environment Variables

| **Name**            | **Default**                                                                | **Required** |
|---------------------|----------------------------------------------------------------------------|--------------|
| `PROVIDER_NAME`     | `aws`                                                                      | false        |
| `SECRET_ID`         |                                                                            | true         |
| `TEMPLATE_BASE64`   | `e3sgcmFuZ2UgJGssICR2IDo9IC4gfX1be3sgJGsgfX1dCnt7ICR2IH19Cgp7eyBlbmQgfX0K` | false        |
| `TEMPLATE_FILENAME` |                                                                            | false        |
| `OUTPUT_FILENAME`   | `output`                                                                   | false        |