package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"winrm/debugger"
	"winrm/mspsrp"
	"winrm/pwshxml"

	"github.com/google/uuid"
	"github.com/oiweiwei/go-msrpc/ssp/credential"
	"github.com/oiweiwei/go-msrpc/ssp/ntlm"
)

// WinRMShell represents a WinRM PowerShell session with NTLM authentication
type WinRMShell struct {
	url        string
	cred       credential.Password
	sessionID  string
	shellID    string
	runspaceID string
	pipelineID string
	client     *http.Client
	auth       *ntlm.Authentifier
}

// NewWinRMShell creates a new WinRM shell instance with NTLM credentials
func NewWinRMShell(url, domain, username, password string) *WinRMShell {
	return &WinRMShell{
		url:        url,
		cred:       credential.NewFromPassword(username, password, credential.Domain(domain)),
		sessionID:  uuid.New().String(),
		runspaceID: uuid.New().String(),
		pipelineID: uuid.New().String(),
		client:     &http.Client{},
	}
}

// CreateShell sends a Create request to establish a PowerShell runspace with NTLM auth
func (w *WinRMShell) CreateShell() error {
	// Step 1: Generate the creationXML using your function
	creationXML, err := pwshxml.CreateCreationXML(w.runspaceID, w.pipelineID)
	if err != nil {
		return fmt.Errorf("failed to create creationXML: %w", err)
	}

	// Step 2: Build the SOAP request
	messageId := uuid.New().String()
	soapRequest, err := mspsrp.EnvelopeToString(mspsrp.CreateShellRequest(creationXML, messageId, w.sessionID))
	if err != nil {
		return fmt.Errorf("failed to create CreateShell XML: %w", err)
	}

	debugger.Println("Creation Shell Request")
	debugger.Println("----------------------------")
	debugger.Println(soapRequest)
	debugger.Println("----------------------------")
	// Step 3: Send with NTLM authentication
	responseBody, err := w.sendWithNTLM([]byte(soapRequest))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	debugger.Println("Creation Shell Response")
	debugger.Println("----------------------------")
	debugger.Println(responseBody)
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")

	// Step 4: Parse ShellID from response using your parser
	shellID, err := mspsrp.ParseShellID(responseBody)
	if err != nil {
		return fmt.Errorf("failed to parse ShellID: %w", err)
	}

	w.shellID = shellID
	fmt.Printf("Shell created successfully! ShellID: %s\n", w.shellID)
	return nil
}

// sendWithNTLM performs NTLM authentication and sends request
func (w *WinRMShell) sendWithNTLM(body []byte) ([]byte, error) {
	ctx := context.Background()
	
	// First request: establish auth if not already done
	if w.auth == nil {
		// Create NTLM config with encryption/signing
		cfg := &ntlm.Config{
			Credential:      w.cred,
			Integrity:       true,
			Confidentiality: true,
		}
		w.auth = &ntlm.Authentifier{Config: cfg}

		// Perform NTLM handshake
		if err := w.performNTLMHandshake(); err != nil {
			return nil, err
		}
	}

	// Now send encrypted request
	return w.sendEncrypted(ctx, body)
}

// performNTLMHandshake does the 3-way NTLM handshake
func (w *WinRMShell) performNTLMHandshake() error {
	ctx := context.Background()

	// Step 1: Trigger 401
	req, _ := http.NewRequest("POST", w.url, nil)
	req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Step 2: Send NTLM Type 1
	neg, err := w.auth.Negotiate(ctx)
	if err != nil {
		return fmt.Errorf("NTLM negotiate: %w", err)
	}

	req, _ = http.NewRequest("POST", w.url, nil)
	req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(neg))
	resp, err = w.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Step 3: Parse challenge
	authHeader := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(authHeader, "Negotiate ") {
		return fmt.Errorf("no NTLM challenge")
	}

	challenge, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Negotiate "))
	if err != nil {
		return fmt.Errorf("decode challenge: %w", err)
	}

	// Step 4: Send NTLM Type 3
	auth, err := w.auth.Authenticate(ctx, challenge)
	if err != nil {
		return fmt.Errorf("NTLM authenticate: %w", err)
	}

	req, _ = http.NewRequest("POST", w.url, nil)
	req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(auth))
	resp, err = w.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("NTLM auth failed: %d", resp.StatusCode)
	}

	return nil
}

// sendEncrypted sends an encrypted multipart request
func (w *WinRMShell) sendEncrypted(ctx context.Context, body []byte) ([]byte, error) {
	protocol := "application/HTTP-SPNEGO-session-encrypted"
	var buf bytes.Buffer

	// Chunk body into 16KB pieces like winrmexec.py
	chunkSize := 16384
	for offset := 0; offset < len(body); offset += chunkSize {
		end := offset + chunkSize
		if end > len(body) {
			end = len(body)
		}
		chunk := body[offset:end]

		// Sign PLAINTEXT first (Python signs plaintext, not ciphertext)
		checksum, err := w.auth.MakeOutboundChecksum(ctx, [][]byte{chunk})
		if err != nil {
			return nil, fmt.Errorf("checksum failed: %w", err)
		}
		sig, err := w.auth.MakeOutboundSignature(ctx, checksum)
		if err != nil {
			return nil, fmt.Errorf("sign failed: %w", err)
		}
		debugger.Println(fmt.Sprintf("[ENCRYPT] Chunk %d: plaintext=%d bytes, sig=%d bytes", offset/chunkSize, len(chunk), len(sig)))

		// Then encrypt chunk
		encrypted := make([]byte, len(chunk))
		copy(encrypted, chunk)
		err = w.auth.ApplyOutboundCipher(ctx, encrypted)
		if err != nil {
			return nil, fmt.Errorf("encrypt failed: %w", err)
		}

		// Build multipart EXACTLY like Python: data += b"Content-Type: ...\r\n" + pack(...) + sig + enc
		buf.WriteString("--Encrypted Boundary\r\n")
		buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", protocol))
		buf.WriteString(fmt.Sprintf("OriginalContent: type=application/soap+xml;charset=UTF-8;Length=%d\r\n", len(chunk)))
		buf.WriteString("--Encrypted Boundary\r\n")
		
		// Python: data += b"Content-Type: application/octet-stream\r\n" + pack("<I", len(sig)) + sig + enc
		// NO blank line between header and binary!
		buf.WriteString("Content-Type: application/octet-stream\r\n")
		sigLen := make([]byte, 4)
		binary.LittleEndian.PutUint32(sigLen, uint32(len(sig)))
		buf.Write(sigLen)
		buf.Write(sig)
		buf.Write(encrypted)
	}

	// Python: data += b"--Encrypted Boundary--\r\n" after loop (with implicit CRLF before it from previous write)
	buf.WriteString("\r\n--Encrypted Boundary--\r\n")

	// Debug: show first 200 bytes of multipart
	debugger.Println(fmt.Sprintf("[ENCRYPT] Multipart length: %d", buf.Len()))
	preview := buf.Bytes()
	if len(preview) > 200 {
		preview = preview[:200]
	}
	debugger.Println(fmt.Sprintf("[ENCRYPT] First 200 bytes: %q", string(preview)))

	// Send encrypted request
	req, _ := http.NewRequest("POST", w.url, &buf)
	req.Header.Set("Content-Type", fmt.Sprintf(`multipart/x-multi-encrypted;protocol="%s";boundary="Encrypted Boundary"`, protocol))
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	debugger.Println(fmt.Sprintf("[ENCRYPT] Response status: %d", resp.StatusCode))
	if resp.StatusCode == 400 {
		// Read error body
		body, _ := io.ReadAll(resp.Body)
		debugger.Println(fmt.Sprintf("[ENCRYPT] 400 Error body: %q", string(body)))
		return nil, fmt.Errorf("HTTP 400: %s", string(body))
	}

	// Parse encrypted response
	return w.decryptResponse(ctx, resp)
}

// decryptResponse decrypts a multipart encrypted response
func (w *WinRMShell) decryptResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "multipart/encrypted") {
		// Not encrypted, read directly
		return io.ReadAll(resp.Body)
	}

	_, params, _ := mime.ParseMediaType(ct)
	boundary := params["boundary"]
	
	mr := multipart.NewReader(resp.Body, boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		if part.Header.Get("Content-Type") == "application/octet-stream" {
			data, _ := io.ReadAll(part)
			
			// Parse signature length
			if len(data) < 4 {
				continue
			}
			
			sigLen := binary.LittleEndian.Uint32(data[:4])
			if len(data) < int(4+sigLen) {
				continue
			}
			
			sig := data[4 : 4+sigLen]
			encrypted := data[4+sigLen:]
			
			// Verify signature by computing expected signature
			checksum, err := w.auth.MakeInboundChecksum(ctx, [][]byte{encrypted})
			if err != nil {
				return nil, fmt.Errorf("checksum failed: %w", err)
			}
			
			expectedSig, err := w.auth.MakeInboundSignature(ctx, checksum)
			if err != nil {
				return nil, fmt.Errorf("signature generation failed: %w", err)
			}
			
			if !bytes.Equal(sig, expectedSig) {
				return nil, fmt.Errorf("signature verification failed: mismatch")
			}
			
			// Decrypt
			decrypted := make([]byte, len(encrypted))
			copy(decrypted, encrypted)
			err = w.auth.ApplyInboundCipher(ctx, decrypted)
			if err != nil {
				return nil, fmt.Errorf("decryption failed: %w", err)
			}
			
			return decrypted, nil
		}
	}
	
	return nil, fmt.Errorf("no encrypted data found")
}


// ExecuteCommand sends a command to the shell and returns the CommandID
func (w *WinRMShell) ExecuteCommand(command string) (string, error) {
	if w.shellID == "" {
		return "", fmt.Errorf("shell not created, call CreateShell() first")
	}

	// Step 1: Generate command arguments using your function
	commandArgs, err := pwshxml.CreateCommandArguments(command, w.runspaceID, w.pipelineID)
	if err != nil {
		return "", fmt.Errorf("failed to create command arguments: %w", err)
	}

	// Step 2: Build the SOAP request
	messageId := uuid.New().String()
	soapRequest, err := mspsrp.EnvelopeToString(mspsrp.CreateCommandRequest(commandArgs, w.shellID, w.sessionID, messageId))
	if err != nil {
		return "", fmt.Errorf("failed to create CreateCommand XML: %w", err)
	}
	debugger.Println("Execute Command Request")
	debugger.Println("----------------------------")
	debugger.Println(soapRequest)
	debugger.Println("----------------------------")

	// Step 3: Send with NTLM authentication
	responseBody, err := w.sendWithNTLM([]byte(soapRequest))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	debugger.Println("Execute Command Response")
	debugger.Println("----------------------------")
	debugger.Println(responseBody)
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")

	// Step 4: Parse CommandID from response
	commandID, err := mspsrp.ParseCommandID(responseBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse CommandID: %w", err)
	}

	fmt.Printf("Command sent! CommandID: %s\n", commandID)
	return commandID, nil
}


// ReceiveOutput retrieves the output from a command
func (w *WinRMShell) ReceiveOutput(commandID string) ([]string, error) {
	// Step 1: Build the SOAP request
	messageId := uuid.New().String()
	soapRequest, err := mspsrp.EnvelopeToString(mspsrp.CreateReceiveRequest(commandID, w.shellID, messageId, w.sessionID))
	if err != nil {
		return nil, fmt.Errorf("failed to create Receive XML: %w", err)
	}
	debugger.Println("Receive Request")
	debugger.Println("----------------------------")
	debugger.Println(soapRequest)
	debugger.Println("----------------------------")

	// Step 2: Send with NTLM authentication
	responseBody, err := w.sendWithNTLM([]byte(soapRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	debugger.Println("Receive Response")
	debugger.Println("----------------------------")
	debugger.Println(responseBody)
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")

	// Step 3: Parse output streams
	stdout, stderr, exitCode, done, err := mspsrp.ParseReceiveOutput(responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	fmt.Printf("Command done: %v, Exit code: %d\n", done, exitCode)
	fmt.Printf("Stdout streams: %d, Stderr streams: %d\n", len(stdout), len(stderr))

	return stdout, nil
}
