# Ledger file
Sigrun also creates a ledger file commonly named `sigrun-ledger.json` which stores a record of every signature made by sigrun in this repository.

```json
{
    "Id":4,
    "Git":{
       "Hash":"723f99d2d1bfb67c882946a0edd05d7148880714",
       "Message":"added ledger signature in .sigrun/ledger.sig\n",
       "Author":"Shravan Shetty \u003cshravanshetty322@gmail.com\u003e",
       "UnixTime":1630842059000000000
    },
    "Hash":"1gMbBfJQpsLSSDKNQMtQWUjMZBJNf8zcLweooAcr4ZE=",
    "Timestamp":"1630844319776051088",
    "Annotations":{
       
    },
    "Checksum":{
       
    }
 }
```

The ledger file contains a series of ledger entires, above is an example of a ledger entry.
### Id
Uniquely identifies a ledger entry
### Git
Various information about the current git commit such as hash, message, author and timestamp.
### Hash
Its hash of all the files in the current repository.
### Timestamp
Timestamp of when this signature was made.
### Checksum
The hash of every file or directory of the repo when the ledger entry was created.