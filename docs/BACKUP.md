# Backing up existing sigrun cluster configuration

To backup the existing configuration, save the config map named `sigrun-controller-config`. 
To load this backup at a later time, deploy the saved config map after sigrun has been installed.

This can also be used to transfer configuration from one cluster to another easily!

You can backup the config map into a file using this command
```
kubectl get configmap sigrun-controller-config -o yaml > sigrun-controller-config.yaml
```

This file will be your backup. To restore your backup run the following command in the new cluster after sigrun has been initialized.
```
kubectl apply -f sigrun-controller-config.yaml
```

