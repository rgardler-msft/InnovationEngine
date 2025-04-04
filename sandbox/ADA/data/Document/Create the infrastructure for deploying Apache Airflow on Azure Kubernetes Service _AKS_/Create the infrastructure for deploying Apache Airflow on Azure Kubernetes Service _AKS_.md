# Create the infrastructure for deploying Apache Airflow on Azure Kubernetes Service (AKS)

## Overview

This document outlines the infrastructure required to deploy Apache Airflow on Azure Kubernetes Service (AKS). Apache Airflow is a powerful workflow orchestration tool used for automating complex data pipelines and workflows. By leveraging AKS, users can deploy Airflow in a scalable, secure, and highly available environment. This guide is tailored for scenarios where users need to deploy Airflow using Helm, ensuring efficient containerized deployment and management.

Use cases for this architecture include orchestrating ETL processes, managing machine learning workflows, and automating data pipeline execution in a cloud-native environment. The document provides step-by-step instructions for setting up the necessary Azure resources, including identity management, storage, container registry, and Kubernetes cluster configuration.

### Major Components

* **Resource Group**: The resource group serves as the logical container for all Azure resources related to the Airflow deployment. It simplifies resource management and provides a unified billing structure.
* **Azure Key Vault**: Used to securely store secrets, such as Airflow passwords and sensitive configuration data. The External Secrets Operator accesses these secrets during runtime.
* **Azure Container Registry (ACR)**: Hosts the container images required for Apache Airflow and its dependencies. ACR ensures secure and efficient image management.
* **Azure Storage Account**: Provides storage for Airflow logs, enabling persistent and centralized logging for troubleshooting and monitoring.
* **Azure Kubernetes Service (AKS)**: Hosts the Apache Airflow deployment. AKS provides a managed Kubernetes environment with features like workload identity, OIDC issuer, and auto-scaling.
* **Managed Identity**: A user-assigned managed identity is created to allow secure access to Azure Key Vault secrets from the AKS cluster.
* **Helm**: A package manager for Kubernetes, used to deploy and manage Apache Airflow on AKS.

### Decision Points

In this architecture, users will encounter several decision points that can impact the deployment's performance, cost, and scalability. Below are the key areas where decisions need to be made:

1. **Resource Group Location**: Choose the Azure region where the resource group will be created. Factors to consider include proximity to data sources, compliance requirements, and latency.
2. **Azure Key Vault Configuration**: Decide whether to enable RBAC authorization for Key Vault. RBAC provides fine-grained access control but may require additional configuration.
3. **Azure Container Registry SKU**: Select the appropriate SKU (e.g., Basic, Standard, Premium) based on the expected image storage size and throughput requirements.
4. **Azure Storage Account SKU**: Choose a storage SKU (e.g., Standard_LRS, Standard_ZRS) based on redundancy and performance needs. Standard_ZRS offers higher availability.
5. **AKS Cluster Configuration**:
   - **Node VM Size**: Select the VM size for AKS nodes based on workload requirements. For example, Standard_DS4_v2 offers a balance of CPU and memory.
   - **Node Count**: Determine the number of nodes required for the cluster. This depends on workload concurrency and expected traffic.
   - **Auto-Upgrade Channel**: Decide whether to enable automatic upgrades for Kubernetes versions and node images.
   - **Network Plugin**: Choose between Azure CNI or Kubenet for networking. Azure CNI is recommended for advanced networking features.
6. **Container Images**: Decide whether to use public Docker Hub images or import them into a private Azure Container Registry for enhanced security.
7. **Helm Chart Configuration**: Configure Helm values for Airflow deployment, including resource limits, replica counts, and environment variables.

Each decision should be guided by workload requirements, cost considerations, and operational preferences. Additional details for these decision points will be provided in the respective sections of the document.

### Alternatives

While this document focuses on deploying Apache Airflow on AKS, there are alternative Azure-based solutions that users may consider:

1. **Azure App Service with Docker**: For simpler deployments, users can run Apache Airflow as a containerized application on Azure App Service. This approach reduces operational complexity but may lack the scalability and flexibility of AKS.
2. **Azure Batch**: For orchestrating large-scale batch processing workflows, Azure Batch can be an alternative to Airflow. It is optimized for parallel compute jobs but may not provide the same level of workflow orchestration.
3. **Azure Data Factory**: For data integration and ETL workflows, Azure Data Factory offers a managed service with built-in connectors and a visual interface. However, it is less customizable compared to Airflow.
4. **Self-Managed Kubernetes**: Users can deploy Apache Airflow on a self-managed Kubernetes cluster hosted on Azure VMs. This provides full control over the Kubernetes environment but requires significant operational overhead.

Each alternative has trade-offs in terms of scalability, cost, and operational complexity. AKS is recommended for users seeking a managed Kubernetes solution with integrated Azure services and Helm-based deployment capabilities.

## Prerequisites

The following prerequisites are required before you are able to work through this document.

- Az CLI is installed and you are logged in to an active Azure subscription

```bash
export SUFFIX=$(date +%s%N | sha256sum | head -c 6)
```

---

## Step 1: Create Resource Group

The resource group serves as the logical container for all Azure resources related to the Airflow deployment. It simplifies resource management and provides a unified billing structure.

Define the environment variables for the resource group:

```bash
export RESOURCE_GROUP_NAME_ED432="AirflowRG_$SUFFIX"
export REGION_ED432="westus2"
```

Create the resource group:

```bash
az group create --name $RESOURCE_GROUP_NAME_ED432 \
    --location $REGION_ED432
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/AirflowRG_xxxxxx",
    "location": "westus2",
    "managedBy": null,
    "name": "AirflowRG_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    },
    "tags": null,
    "type": "Microsoft.Resources/resourceGroups"
}
```

The choice of region impacts latency, compliance, and cost. WestUS2 is recommended for its availability and proximity to many data centers.

---

## Step 2: Create Azure Key Vault

Azure Key Vault securely stores secrets, such as Airflow passwords and sensitive configuration data. The External Secrets Operator accesses these secrets during runtime.

Define the environment variables for the Key Vault:

```bash
export KEY_VAULT_NAME_ED432="airflowkv$SUFFIX"
```

Create the Key Vault:

```bash
az keyvault create --name $KEY_VAULT_NAME_ED432 \
    --resource-group $RESOURCE_GROUP_NAME_ED432 \
    --location $REGION_ED432 \
    --sku standard
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/AirflowRG_xxxxxx/providers/Microsoft.KeyVault/vaults/airflowkvxxxxxx",
    "location": "westus2",
    "name": "airflowkvxxxxxx",
    "properties": {
        "sku": {
            "family": "A",
            "name": "standard"
        },
        "vaultUri": "https://airflowkvxxxxxx.vault.azure.net/"
    },
    "type": "Microsoft.KeyVault/vaults"
}
```

Consider enabling RBAC authorization for finer-grained access control, especially in production environments.

---

## Step 3: Create Azure Container Registry (ACR)

Azure Container Registry hosts the container images required for Apache Airflow and its dependencies. ACR ensures secure and efficient image management.

Define the environment variables for the ACR:

```bash
export ACR_NAME_ED432="airflowacr$SUFFIX"
```

Create the ACR:

```bash
az acr create --name $ACR_NAME_ED432 \
    --resource-group $RESOURCE_GROUP_NAME_ED432 \
    --location $REGION_ED432 \
    --sku Standard
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/AirflowRG_xxxxxx/providers/Microsoft.ContainerRegistry/registries/airflowacrxxxxxx",
    "location": "westus2",
    "name": "airflowacrxxxxxx",
    "properties": {
        "sku": {
            "name": "Standard"
        },
        "loginServer": "airflowacrxxxxxx.azurecr.io"
    },
    "type": "Microsoft.ContainerRegistry/registries"
}
```

The Standard SKU is suitable for most workloads. Consider upgrading to Premium for higher throughput and geo-replication.

---

## Step 4: Create Azure Storage Account

Azure Storage Account provides storage for Airflow logs, enabling persistent and centralized logging for troubleshooting and monitoring.

Define the environment variables for the storage account:

```bash
export STORAGE_ACCOUNT_NAME_ED432="airflowstorage$SUFFIX"
```

Create the storage account:

```bash
az storage account create --name $STORAGE_ACCOUNT_NAME_ED432 \
    --resource-group $RESOURCE_GROUP_NAME_ED432 \
    --location $REGION_ED432 \
    --sku Standard_LRS
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/AirflowRG_xxxxxx/providers/Microsoft.Storage/storageAccounts/airflowstoragexxxxxx",
    "location": "westus2",
    "name": "airflowstoragexxxxxx",
    "properties": {
        "provisioningState": "Succeeded",
        "primaryEndpoints": {
            "blob": "https://airflowstoragexxxxxx.blob.core.windows.net/"
        }
    },
    "sku": {
        "name": "Standard_LRS"
    },
    "type": "Microsoft.Storage/storageAccounts"
}
```

Standard_LRS is cost-effective and suitable for most workloads. Consider Standard_ZRS for higher availability.

---

## Step 5: Create Azure Kubernetes Service (AKS)

Azure Kubernetes Service hosts the Apache Airflow deployment. AKS provides a managed Kubernetes environment with features like workload identity, OIDC issuer, and auto-scaling.

Define the environment variables for the AKS cluster:

```bash
export AKS_CLUSTER_NAME_ED432="AirflowAKS$SUFFIX"
export NODE_VM_SIZE_ED432="Standard_D4_v3"
export NODE_COUNT_ED432=3
```

Create the AKS cluster:

```bash
az aks create --name $AKS_CLUSTER_NAME_ED432 \
    --resource-group $RESOURCE_GROUP_NAME_ED432 \
    --location $REGION_ED432 \
    --node-vm-size $NODE_VM_SIZE_ED432 \
    --node-count $NODE_COUNT_ED432 \
    --enable-managed-identity \
    --enable-oidc-issuer \
    --network-plugin azure
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/AirflowRG_xxxxxx/providers/Microsoft.ContainerService/managedClusters/AirflowAKSxxxxxx",
    "location": "westus2",
    "name": "AirflowAKSxxxxxx",
    "properties": {
        "provisioningState": "Succeeded",
        "kubernetesVersion": "1.25.6",
        "nodeResourceGroup": "MC_AirflowRG_xxxxxx_AirflowAKSxxxxxx_westus2"
    },
    "type": "Microsoft.ContainerService/managedClusters"
}
```

Standard_D4_v3 is a balanced VM size offering good performance and availability in WestUS2. Adjust node count and VM size based on workload requirements.

---

## Step 6: Deploy Apache Airflow using Helm

Helm is a package manager for Kubernetes, used to deploy and manage Apache Airflow on AKS.

Define the environment variables for Helm deployment:

```bash
export AIRFLOW_RELEASE_NAME_ED432="airflow"
export AIRFLOW_NAMESPACE_ED432="airflow"
```

Create the namespace and deploy Airflow:

```bash
kubectl create namespace $AIRFLOW_NAMESPACE_ED432

helm repo add apache-airflow https://airflow.apache.org

helm install $AIRFLOW_RELEASE_NAME_ED432 apache-airflow/airflow \
    --namespace $AIRFLOW_NAMESPACE_ED432 \
    --set executor=CeleryExecutor \
    --set logs.persistence.enabled=true \
    --set logs.persistence.storageClassName=default \
    --set logs.persistence.size=10Gi
```

This command will output results similar to the following:

```text
NAME: airflow
LAST DEPLOYED: Mon Oct 16 14:00:00 2023
NAMESPACE: airflow
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Ensure that `kubelogin` is installed and configured correctly to avoid authentication issues. Customize Helm values for resource limits, replica counts, and environment variables based on workload requirements.

## Summary

This document detailed the architecture and step-by-step process for deploying a simple Python Flask application to Microsoft Azure using Azure App Service. The purpose of the architecture was to provide a lightweight, fully managed solution for hosting a Flask-based web application, ideal for small-scale projects, prototypes, or educational use cases. By leveraging Azure App Service, developers were able to focus on application development without the need to manage underlying infrastructure.

Key decision points included selecting an App Service Plan SKU, choosing an Azure region, configuring scaling options, determining deployment methods, and evaluating the need for persistent storage. The Free tier of the App Service Plan was recommended for demonstration purposes, while higher tiers were suggested for production workloads. Alternatives such as Azure Kubernetes Service, Azure Functions, and Virtual Machines were discussed for scenarios requiring more control, scalability, or customization.

### Next Steps

* **Test the Application** - Verify the functionality of the deployed Flask application by accessing its URL and ensuring it operates as expected.
* **Enable Monitoring** - Set up Azure Monitor or Application Insights to track performance, availability, and cost metrics for the deployed application.
* **Optimize for Production** - Evaluate scaling options, upgrade the App Service Plan tier, and integrate CI/CD pipelines for streamlined production deployments.

