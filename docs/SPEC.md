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

### Basic Flow
Please refer to [this](./USAGE.md).

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


### CLI Commands

#### Producer related

##### kubectl sigrun init
Creates the config file 

##### kubectl sigrun sign
Signs the list of images in the config file and pushes the signature to the container registry

#### Consumer related

##### kubectl sigrun add
Parses the config file from the consumer and updates the policy agent to verify the listed images

##### kubectl sigrun list
Lists the current config files that have been added with metadata about them.

##### kubectl sigrun remove
Removes a config file from the policy agent which should remove the list of images and metadata related to the config file.