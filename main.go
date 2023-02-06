package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
		if err != nil {
			return err
		}

		// Create an Azure resource (Storage Account)
		account, err := storage.NewStorageAccount(ctx, "sa", &storage.StorageAccountArgs{
			ResourceGroupName: resourceGroup.Name,
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Standard_LRS"),
			},
			Kind: pulumi.String("StorageV2"),
		})
		if err != nil {
			return err
		}

		// Enable static website support
		staticWebsite, err := storage.NewStorageAccountStaticWebsite(ctx, "staticWebsite", &storage.StorageAccountStaticWebsiteArgs{
			AccountName:       account.Name,
			ResourceGroupName: resourceGroup.Name,
			IndexDocument:     pulumi.String("index.html"),
		})
		if err != nil {
			return err
		}

		// Upload the file
		_, err = storage.NewBlob(ctx, "index.html", &storage.BlobArgs{
			ResourceGroupName: resourceGroup.Name,
			AccountName:       account.Name,
			ContainerName:     staticWebsite.ContainerName,
			Source:            pulumi.NewFileAsset("index.html"),
			ContentType:       pulumi.String("text/html"),
		})
		if err != nil {
			return err
		}

		// Export the primary key of the Storage Account
		ctx.Export("primaryStorageKey", pulumi.All(resourceGroup.Name, account.Name).ApplyT(
			func(args []interface{}) (string, error) {
				resourceGroupName := args[0].(string)
				accountName := args[1].(string)
				accountKeys, err := storage.ListStorageAccountKeys(ctx, &storage.ListStorageAccountKeysArgs{
					ResourceGroupName: resourceGroupName,
					AccountName:       accountName,
				})
				if err != nil {
					return "", err
				}

				return accountKeys.Keys[0].Value, nil
			},
		))

		ctx.Export("staticEndpoint", account.PrimaryEndpoints.Web())

		return nil
	})
}
