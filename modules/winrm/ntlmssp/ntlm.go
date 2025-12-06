package ntlmssp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/bodgit/windows"
	"golang.org/x/crypto/md4"
)

const (
	lmCiphertext string = "KGS!@#$%"
)

func realCurrentTime() ([]byte, error) {
	ft := windows.NsecToFiletime(time.Now().UnixNano())

	b := bytes.Buffer{}
	if err := binary.Write(&b, binary.LittleEndian, ft); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

var currentTime = realCurrentTime

func LmowfV1(password string) ([]byte, error) {
	b := bytes.Buffer{}

	padded := zeroPad([]byte(strings.ToUpper(password)), 14)

	for _, i := range []int{0, 7} {
		result, err := encryptDES(padded[i:], []byte(lmCiphertext))
		if err != nil {
			return nil, err
		}

		b.Write(result)
	}

	return b.Bytes(), nil
}

func NtowfV1(password string) ([]byte, error) {
	b, err := utf16FromString(password)
	if err != nil {
		return nil, err
	}
	return hashMD4(b), nil
}

func ntowfV2(username, ntHash, domain string) ([]byte, error) {
	m, err := utf16FromString(strings.ToUpper(username) + domain)
	if err != nil {
		return nil, err
	}

	k, err := hex.DecodeString(ntHash)
	if err != nil {
		return nil, err
	}

	return hmacMD5(k, m), nil
}

func lmV1WithSessionSecurityResponse(clientChallenge []byte) []byte {
	return zeroPad(clientChallenge, 24)
}

func lmV1Response(lmHash string, serverChallenge []byte) ([]byte, error) {
	lmByte, err := hex.DecodeString(lmHash)
	if err != nil {
		return nil, err
	}
	return encryptDESL(lmByte, serverChallenge)
}

func lmV2Response(username, ntlmHash, domain string, serverChallenge, clientChallenge []byte) ([]byte, error) {
	ntlmHashV2, err := ntowfV2(username, ntlmHash, domain)
	if err != nil {
		return nil, err
	}
	return concat(hmacMD5(ntlmHashV2, concat(serverChallenge, clientChallenge)), clientChallenge), nil
}

func ntlmV1Response(ntlmHash string, serverChallenge []byte) ([]byte, []byte, error) {
	ntByte, err := hex.DecodeString(ntlmHash)
	if err != nil {
		return nil, nil, err
	}
	response, err := encryptDESL(ntByte, serverChallenge)
	if err != nil {
		return nil, nil, err
	}

	return response, hashMD4([]byte(ntByte)), nil
}

func ntlm2Response(ntlmHash string, serverChallenge, clientChallenge []byte) ([]byte, []byte, error) {

	ntByte, err := hex.DecodeString(ntlmHash)
	if err != nil {
		return nil, nil, err
	}
	response, err := encryptDESL([]byte(ntByte), hashMD5(concat(serverChallenge, clientChallenge))[:8])
	if err != nil {
		return nil, nil, err
	}

	return response, hashMD4([]byte(ntByte)), nil
}

func ntlmV2Temp(timestamp []byte, clientChallenge []byte, targetInfo targetInfo) ([]byte, error) {
	b, err := targetInfo.Marshal()
	if err != nil {
		return nil, err
	}

	return concat([]byte{0x01}, []byte{0x01}, zeroBytes(6), timestamp, clientChallenge, zeroBytes(4), b, zeroBytes(4)), nil
}

func ntlmV2Response(username, nthash, domain string, serverChallenge, clientChallenge []byte, timestamp []byte, targetInfo targetInfo) ([]byte, []byte, error) {
	ntlmHash, err := ntowfV2(username, nthash, domain)
	if err != nil {
		return nil, nil, err
	}

	temp, err := ntlmV2Temp(timestamp, clientChallenge, targetInfo)
	if err != nil {
		return nil, nil, err
	}

	ntProofStr := hmacMD5(ntlmHash, concat(serverChallenge, temp))

	return concat(ntProofStr, temp), hmacMD5(ntlmHash, ntProofStr), nil
}

func ntlmV1ExchangeKey(flags uint32, sessionBaseKey []byte, serverChallenge []byte, lmChallengeResponse []byte, lmHash []byte) ([]byte, error) {
	switch {
	case ntlmsspNegotiateExtendedSessionsecurity.IsSet(flags):
		return hmacMD5(sessionBaseKey, concat(serverChallenge, lmChallengeResponse[:8])), nil
	case ntlmsspNegotiateLMKey.IsSet(flags):
		b := bytes.Buffer{}

		for _, k := range [][]byte{lmHash[:7], concat(lmHash[7:8], bytes.Repeat([]byte{0xbd}, 6))} {
			result, err := encryptDES(k, lmChallengeResponse[:8])
			if err != nil {
				return nil, err
			}

			b.Write(result)
		}

		return b.Bytes(), nil
	case ntlmsspRequestNonNTSessionKey.IsSet(flags):
		return zeroPad(lmHash[:8], 16), nil
	default:
		return sessionBaseKey, nil
	}
}

func lmChallengeResponse(flags uint32, level lmCompatibilityLevel, clientChallenge []byte, username, hash, domain string, cm *challengeMessage) ([]byte, error) {
	splitted := strings.Split(hash, ":")
	if len(splitted) != 2 {
		return nil, fmt.Errorf("Error in hash format, should be LM:NT")
	}
	lm := splitted[0]
	nt := splitted[1]
	switch {
	case ntlmsspAnonymous.IsSet(flags):
		return zeroBytes(1), nil
	case ntlmsspNegotiateExtendedSessionsecurity.IsSet(flags) && level < 3:
		// LMv1 with session security
		return lmV1WithSessionSecurityResponse(clientChallenge), nil
	case level < 2:
		// LMv1 response
		return lmV1Response(lm, cm.ServerChallenge[:])
	case level == 2:
		// NTLMv1 response
		response, _, err := ntlmV1Response(nt, cm.ServerChallenge[:])
		return response, err
	default:
		// LMv2 response
		if _, ok := cm.TargetInfo.Get(msvAvTimestamp); ok {
			return zeroBytes(24), nil
		}
		return lmV2Response(username, nt, domain, cm.ServerChallenge[:], clientChallenge)
	}
}

func ntChallengeResponse(flags uint32, level lmCompatibilityLevel, clientChallenge []byte, username, hash, domain string, cm *challengeMessage, lmChallengeResponse []byte, targetInfo targetInfo, channelBindings *ChannelBindings) ([]byte, []byte, error) {
	splitted := strings.Split(hash, ":")
	if len(splitted) != 2 {
		return nil, nil, fmt.Errorf("Error in hash format, should be LM:NT")
	}
	lm := splitted[0]
	nt := splitted[1]
	switch {
	case ntlmsspAnonymous.IsSet(flags):
		return []byte{}, zeroBytes(md4.Size), nil
	case level < 3:
		var response, sessionBaseKey []byte
		var err error

		if ntlmsspNegotiateExtendedSessionsecurity.IsSet(flags) {
			// NTLMv1 authentication with NTLM2
			response, sessionBaseKey, err = ntlm2Response(nt, cm.ServerChallenge[:], clientChallenge)
		} else {
			// NTLMv1 authentication
			response, sessionBaseKey, err = ntlmV1Response(nt, cm.ServerChallenge[:])
		}
		if err != nil {
			return nil, nil, err
		}

		lmByte, err := hex.DecodeString(lm)
		if err != nil {
			return nil, nil, err
		}

		keyExchangeKey, err := ntlmV1ExchangeKey(flags, sessionBaseKey, cm.ServerChallenge[:], lmChallengeResponse, lmByte)
		if err != nil {
			return nil, nil, err
		}

		return response, keyExchangeKey, nil
	default:
		// NTLMv2 authentication
		timestamp, ok := targetInfo.Get(msvAvTimestamp)
		if ok {
			var flags uint32
			if v, ok := targetInfo.Get(msvAvFlags); ok {
				flags = binary.LittleEndian.Uint32(v)
				flags |= msvAvFlagMICProvided
			} else {
				flags = msvAvFlagMICProvided
			}
			v := make([]byte, 4)
			binary.LittleEndian.PutUint32(v, flags)
			targetInfo.Set(msvAvFlags, v)
		} else {
			var err error
			timestamp, err = currentTime()
			if err != nil {
				return nil, nil, err
			}
		}

		if channelBindings != nil {
			b, err := channelBindings.marshal()
			if err != nil {
				return nil, nil, err
			}
			targetInfo.Set(msvChannelBindings, hashMD5(b))
		}

		return ntlmV2Response(username, nt, domain, cm.ServerChallenge[:], clientChallenge, timestamp, targetInfo)
	}
}
