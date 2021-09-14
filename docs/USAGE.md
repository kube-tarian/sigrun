### Basic flow

User - Entity that manages a sigrun repo

Consumer - Entity that wants to use docker images created by the users sigrun repo.

1. The user initializes a sigrun repo using "kubectl sigrun init" command inside a folder, the command creates a config file.
The user is asked to enter all the images and the information needed to sign and verify images while running this command.

2. The consumer initializes their cluster to verify container images using "kubectl sigrun init cluster"
command which will add a policy agent to the cluster such as kyverno or opa or a custom controller. The policy agent will be used to verify the images.

3. The consumer adds the list of allowed images from the user by parsing the config file of the sigrun repo 
created by the user using the "kubectl sigrun add 'link to sigrun repo'" command.
This command will create a policy which will add the images and a way to verify each image(pubkey/cert).

4. The consumer try to pull an image from the user and fails since the user has not signed his images.

5. The user signs the latest images and pushes it to the container registry using the
"kubectl sigrun sign" command.

6. The consumer try to pull an image again and succeeds since the image signature has been verified.
