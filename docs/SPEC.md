# Specification

## Who?
There are 2 entities in the system producers and consumers. Producers create container images while consumers use
the image in their clusters.

## What?
The current application allows consumers to only allow verified images into their kubernetes cluster.
It aims to work as a manager for container image verification in a kubernetes cluster.
It also allows producers to seamlessly sign images related to your project.

## Why?
There should be an easy way to seamlessly verify images that are pulled to a cluster.
Current existing solutions can be simplified.

## How?
For a consumer to allow images into their cluster, they will have to specifically add them to the list of allowed images along with 
a method to verify them. This is done by parsing the config file of a sigrun repo which was created 
by a producer. 

The config file contains a list of images container registry paths,
along with a way to verify those images.

The config file will be named "sigrun-config.json" in a sigrun repo.

All tasks related to sigrun are managed by the sigrun CLI command which can be installed by referring to the
repository readme.

### Basic flow

1. The producer creates a sigrun repo using the "kubectl sigrun init" command inside a folder, the command creates a config file
and a directory to store the update chain (will cover this in a later section). 
The producer is asked to enter all the images and the information needed to sign and verify images while running this command.

2. The consumer initializes their cluster to verify container images using "kubectl sigrun init cluster"
command which will add a policy agent to the cluster such as kyverno or opa. The policy agent will be used to verify the images.

3. The consumer adds the list of allowed images from the producer by parsing the config file of the sigrun repo 
created by the producer using the "kubectl sigrun add 'link to sigrun repo'" command.
This command will create a policy which will add the images and a way to verify each image(pubkey/cert).

4. The consumer try to pull an image from the producer and fails since the producer has not signed his images.

5. The producer signs the latest images and pushes it to the container registry using the
"kubectl sigrun sign" command.

6. The consumer try to pull an image again and succeeds since the image signature has been verified.

### Update chain

There can be a scenario where the sigrun config needs to change. Examples of these scenarios include
1. A new image has to be added to the list of images
2. A fresh pair of keys has to be added
3. Some metadata needs to change

and so on...

For these scenarios we want an easy way for producers to update the sigrun config while 
allowing the consumer to verify it.

This needs to be done carefully, the consumer should be able to verify that the config has not been updated by a malicious party.

The "update chain" concept allows the producer to update the config while allowing the consumer to verify it easily.

When the producer wants to update the config,
the producer will update the config file as he wishes and then run "kubectl sigrun configure" command.
The command will sign the new config file with the credentials of the old config file. Hence for a malicious entity to 
update the config file, they will need access to the existing credentials.

The sigrun command creates a ".sigrun" folder during repo initialization which is used to store the "update chain".
At the start the ".sigrun" folder contains only the initial file "0.md", where 0 referring to the first config file
created using "kubectl sigrun init". Every time "kubectl sigrun configure" is run a new file is created in ".sigrun".
The new file will be named "1.md", "2.md" and so on. The current order of the file in the update chain is
recorded in the "ChainNo" field in the config file.

On the consumer end, the consumer can pull any new images added by the producer using the "kubectl sigrun update"
command. This command will pull the latest sigrun config from the sigrun repo. 

If the file has been updated (ChainNo greater than existing), sigrun will start verifying the new config file.
While adding a new config file using "kubectl sigrun add" the "ChainNo" and "PublicKey" fields 
were stored in the cluster. 

Lets assume "ChainNo" was 0 when intially adding the config and the "ChainNo" in the new file is 2.
Sigrun will start verifying the chain by going to ".sigrun/1.md" (initial "ChainNo" + 1) and 
verifying the signature with the public key it received from 0 while adding the config. Once 1.md is verified,
it will verify 2.md using the creds of 1.md and so on.

If the latest file was successfully verified, the latest changes are updated to the cluster.


### Config file

Following is an example of a sigrun config file
```json
{
	"ChainNo": 0,
	"Metadata": {
		"Organization": "clix",
		"Developers": [
			"Chandu",
			"Shravan"
		]
	},
	"PublicKey": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEGB7Vpxn4ioyOmMZo65D+k9uY3tW6\n0BvffR1JBZXjRR3jap1T8uDxP6XEJcLBECbzNPZl1ZVYET85qJVMBeXu2A==\n-----END PUBLIC KEY-----\n",
	"PrivateKey": "-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----\neyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6\nOCwicCI6MX0sInNhbHQiOiJPUDFwUnNYYmNDeWNaUnVpOWRjZUMxQVhqakh1UmVr\nNWlPZFp1WHNScHdFPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94\nIiwibm9uY2UiOiJGNzk3WFJscHU4a0JDOGVyK3R5TkppWldxemEvZ3ZlcSJ9LCJj\naXBoZXJ0ZXh0IjoickF6TzdtT2lkYUhNdTlDd0VkclpSNmtYYmtzclRZQTJadzRv\nWmIwam9QRkxrd3FPSi9XZEk2Z1hKMVJBL3lGaVJKbW5US2s4MUhyK1lvTzRPamlj\neE40UncvTGErQWpURSs2bTZFVVFrbGFjNlduQzJjWlhRUUVpdVV4MGsrWTZOMGFL\nWEF1VXZUenE3aFRnWVM1dk94U1NEcG1objV2RERUQWpVZVZpNkpBL3JCRXQ2T3RX\nckhDQlNRYm01K2t5VnJ1bldCZDQxM09lL0E9PSJ9\n-----END ENCRYPTED COSIGN PRIVATE KEY-----\n",
	"Images": [
		"docker.io/shravanss/sigrun-example"
	],
	"Signature": ""
}

```
#### Images
This is a list of container images that is owned by the producer.

#### PublicKey
This field contains the public key which will be used to verify the signature received by the consumer.

#### PrivateKey
This is the encrypted private key of the producer. This is only stored for convenience in the json file.
Ideally it would be stored in seperate location or managed by a KMS (Key management system).
This private key is used while signing images created by the producer.

#### Metadata
Metadata which is used as annotations in the signature for additional verification.

#### ChainNo
This field specifies the order of this file with respect to its "update chain".

#### Signature
The signature of the previous file in the "update chain". The signature is empty in this case since there was no file
before this since its the initial file.

### CLI Commands

#### Producer related

##### kubectl sigrun init
Creates the config file along with ".sigrun" folder to store
the updatechain

##### kubectl sigrun sign
Signs the list of images in the config file and pushes the signature to the container registry

##### kubectl sigrun configure
Signs the current file with the latest file in the update chain and appends the current file to the update chain. 

#### Consumer related

##### kubectl sigrun add
Parses the config file from the consumer and updates the policy agent to verify the listed images

##### kubectl sigrun list
Lists the current config files that have been added with metadata about them.

##### kubectl sigrun update
Checks the config file from the original link. If the file has been updates, verifies the new file and updates the policy if
the file is verified.

##### kubectl sigrun remove
Removes a config file from the policy agent which should remove the list of images and metadata related to the config file.