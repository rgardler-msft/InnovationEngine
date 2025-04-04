# Manage Azure resource groups by using Azure CLI

## Overview

This document provides a comprehensive guide on managing Azure resource groups using the Azure Command-Line Interface (CLI). Azure resource groups are essential containers that hold related resources for an Azure solution, enabling efficient organization, deployment, and management of resources. The Azure CLI is a powerful tool for interacting with Azure services, allowing users to automate and streamline resource group operations such as creation, listing, deletion, deployment, tagging, locking, and access management. This guide is suitable for developers, IT administrators, and cloud architects who need to manage Azure resource groups programmatically or automate their workflows.

### Major Components

* **Azure CLI** - The primary tool used to execute commands for managing Azure resource groups. Users must install and authenticate the Azure CLI before proceeding.
* **Resource Groups** - Containers for logically organizing Azure resources that share the same lifecycle. Resource groups are central to Azure management and serve as the foundation for deploying and managing resources.
* **Azure Resource Manager** - The underlying service that facilitates the management of Azure resources and resource groups. It ensures consistency and provides features like role-based access control (RBAC) and tagging.

### Decision Points

In this architecture, users will encounter several decision points that can significantly impact the efficiency and organization of their Azure environment. These include:

1. **Resource Group Naming**:
   - Choose meaningful names that reflect the purpose or scope of the resource group.
   - Consider naming conventions for easier identification and management.

2. **Location Selection**:
   - Decide on the region where the resource group metadata will be stored.
   - Factors to consider include compliance requirements, latency, and availability.

3. **Azure CLI Commands**:
   - Select appropriate commands for specific operations (e.g., `az group create`, `az group list`, `az group delete`).
   - Understand command parameters and options for customization.

4. **Resource Deployment**:
   - Determine whether to deploy resources directly using Azure CLI or through ARM templates/Bicep files.
   - Evaluate the complexity of the deployment and the need for repeatability.

5. **Locking and Access Control**:
   - Decide on locking policies to prevent accidental deletion or modification of critical resources.
   - Use Azure RBAC to manage access permissions effectively.

6. **Tagging Strategy**:
   - Develop a tagging strategy to logically organize resources and enable cost tracking or filtering.

7. **Exporting Templates**:
   - Decide whether to export existing resource group configurations to ARM templates for reuse or version control.

### Alternatives

While this document focuses on managing Azure resource groups using Azure CLI, there are alternative approaches that users may consider:

1. **Azure Portal**:
   - Provides a graphical interface for managing resource groups.
   - Suitable for users who prefer a visual approach or are new to Azure.

2. **Azure PowerShell**:
   - Offers similar functionality to Azure CLI but uses PowerShell cmdlets.
   - Ideal for users already familiar with PowerShell scripting.

3. **REST API**:
   - Enables programmatic access to Azure services through HTTP requests.
   - Suitable for advanced automation scenarios or integration with external systems.

4. **Terraform**:
   - A third-party Infrastructure-as-Code (IaC) tool for managing Azure resources.
   - Useful for multi-cloud environments or complex deployments.

Each alternative has its strengths and trade-offs. Azure CLI is particularly well-suited for users who prefer command-line tools and need to automate resource group management tasks efficiently.

## Prerequisites

The following prerequisites are required before you are able to work through this document.

- Az CLI is installed and you are logged in to an active Azure subscription.

```bash
export SUFFIX=$(date +%s%N | sha256sum | head -c 6)
```

---

## Step 1: Create a Resource Group

In this step, we will create an Azure resource group, which serves as a container for logically organizing Azure resources that share the same lifecycle. Resource groups are foundational to Azure management and are required before deploying any resources.

We will define the following environment variables:

- `RESOURCE_GROUP_NAME_ED321`: The name of the resource group. This should be meaningful and reflect the purpose or scope of the group. For example, "WebAppRG_$SUFFIX" for a resource group dedicated to a web application.
- `REGION_ED321`: The Azure region where the resource group metadata will be stored. Common choices include `westus2`, `eastus`, or `centralus`.

```bash
export RESOURCE_GROUP_NAME_ED321="ExampleRG_$SUFFIX"
export REGION_ED321="westus2"

az group create --name $RESOURCE_GROUP_NAME_ED321 \
    --location $REGION_ED321
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ExampleRG_xxxxxx",
    "location": "westus2",
    "managedBy": null,
    "name": "ExampleRG_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    },
    "tags": null,
    "type": "Microsoft.Resources/resourceGroups"
}
```

When choosing a region, consider factors such as compliance requirements, latency, and availability. Some regions may have specific limitations or higher costs associated with certain services. Ensure the region aligns with your workload requirements.

---

## Step 2: List Resource Groups

This step demonstrates how to list all resource groups in your Azure subscription. Listing resource groups is useful for auditing, verifying deployments, or identifying existing groups.

```bash
az group list --output table
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
Name                  Location    Status
--------------------  ----------  ---------
ExampleRG_xxxxxx      westus2     Succeeded
AnotherResourceGroup  eastus      Succeeded
```

The `--output table` option provides a human-readable format. You can also use `--output json` or `--output yaml` for programmatic access to the data.

---

## Step 3: Delete a Resource Group

In this step, we will delete a resource group. Deleting a resource group removes all resources contained within it, so use this command with caution.

We will define the following environment variable:

- `RESOURCE_GROUP_TO_DELETE_ED321`: The name of the resource group to delete.

```bash
export RESOURCE_GROUP_TO_DELETE_ED321="ExampleRG_$SUFFIX"

az group delete --name $RESOURCE_GROUP_TO_DELETE_ED321 --yes
```

This command will output results similar to the following.

```text
{
    "status": "Deleting",
    "name": "ExampleRG_xxxxxx"
}
```

Ensure you have reviewed the contents of the resource group before deletion to avoid accidentally removing critical resources. The `--yes` flag bypasses the confirmation prompt for automation purposes.

---

## Step 4: Add Tags to a Resource Group

Tags are key-value pairs that help organize and categorize Azure resources. In this step, we will add tags to a resource group.

We will define the following environment variables:

- `RESOURCE_GROUP_NAME_ED321`: The name of the resource group to tag.
- `TAG_KEY_ED321`: The key for the tag.
- `TAG_VALUE_ED321`: The value for the tag.

```bash
export TAG_KEY_ED321="Environment"
export TAG_VALUE_ED321="Production"

az group update --name $RESOURCE_GROUP_NAME_ED321 \
    --set tags.$TAG_KEY_ED321=$TAG_VALUE_ED321
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ExampleRG_xxxxxx",
    "location": "westus2",
    "managedBy": null,
    "name": "ExampleRG_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    },
    "tags": {
        "Environment": "Production"
    },
    "type": "Microsoft.Resources/resourceGroups"
}
```

Tags can be used for cost tracking, filtering, or grouping resources logically. Develop a consistent tagging strategy to maximize their utility.

---

## Step 5: Lock a Resource Group

Locks prevent accidental deletion or modification of critical resources. In this step, we will apply a lock to a resource group.

We will define the following environment variables:

- `LOCK_NAME_ED321`: The name of the lock.
- `LOCK_LEVEL_ED321`: The level of the lock. Valid values are `ReadOnly` (prevents modifications) and `CanNotDelete` (prevents deletion).

```bash
export LOCK_NAME_ED321="CriticalLock"
export LOCK_LEVEL_ED321="CanNotDelete"

az lock create --name $LOCK_NAME_ED321 \
    --lock-type $LOCK_LEVEL_ED321 \
    --resource-group $RESOURCE_GROUP_NAME_ED321
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ExampleRG_xxxxxx/providers/Microsoft.Authorization/locks/CriticalLock",
    "level": "CanNotDelete",
    "name": "CriticalLock",
    "notes": null,
    "owners": null,
    "properties": {
        "level": "CanNotDelete"
    },
    "type": "Microsoft.Authorization/locks"
}
```

Locks are essential for protecting critical resources. Use `ReadOnly` for resources that should not be modified and `CanNotDelete` for resources that must not be deleted.

---

This deployment section provides step-by-step guidance for managing Azure resource groups using Azure CLI, including creation, listing, deletion, tagging, and locking. Customize the environment variables and commands as needed to align with your specific workload requirements.

## Summary

This document detailed the architecture and step-by-step process for deploying a simple Python Flask application to Microsoft Azure using Azure App Service. The purpose of the architecture was to provide a lightweight, fully managed solution for hosting a Flask-based web application, ideal for small-scale projects, prototypes, or educational use cases. By leveraging Azure App Service, developers were able to focus on application development without the need to manage underlying infrastructure.

Key decision points included selecting an App Service Plan SKU, choosing an Azure region, configuring scaling options, determining deployment methods, and evaluating the need for persistent storage. The Free tier of the App Service Plan was recommended for demonstration purposes, while higher tiers were suggested for production workloads. Alternatives such as Azure Kubernetes Service, Azure Functions, and Virtual Machines were discussed for scenarios requiring more control, scalability, or customization.

### Next Steps

* **Test the Application** - Verify the functionality of the deployed Flask application by accessing its URL and ensuring it operates as expected.
* **Enable Monitoring** - Set up Azure Monitor or Application Insights to track performance, availability, and cost metrics for the deployed application.
* **Optimize for Production** - Evaluate scaling options, upgrade the App Service Plan tier, and integrate CI/CD pipelines for streamlined production deployments.

