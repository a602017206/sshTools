# Cross-Device Export Passphrase Encryption Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow encrypted export/import to work across devices by using a user-supplied passphrase.

**Architecture:** Add passphrase-based AES-GCM encryption for exported passwords with Argon2id-derived key and a per-export salt. Export JSON includes encryption metadata; import detects metadata and requires passphrase to decrypt, storing passwords in the local credential store.

**Tech Stack:** Go (crypto/aes, cipher.GCM), golang.org/x/crypto/argon2, Svelte (Dialog/InputDialog), Wails bindings.

---

### Task 1: Define passphrase encryption format + helpers

**Files:**
- Modify: `app.go`
- Modify: `internal/store/credentials.go` (optional helper if reused)
- Test: `app_export_passphrase_test.go`

**Step 1: Write the failing test**

Create `app_export_passphrase_test.go` with a test that:
- Sets up two connections with passwords.
- Calls `ExportConnectionsByIDsWithPassphrase(ids, passphrase)`.
- Verifies export JSON includes `password_encryption` metadata with mode `passphrase` and a base64 salt.
- Verifies each password value is prefixed with `penc:` and not plaintext.

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL because export method doesn’t exist.

**Step 3: Implement minimal helpers**

In `app.go`, add:
- `type PasswordEncryption struct { Mode string `json:"mode"`; Salt string `json:"salt"`; KDF string `json:"kdf"` }`
- Extend `ExportData` with `PasswordEncryption *PasswordEncryption `json:"password_encryption,omitempty"``
- Helper `derivePassphraseKey(passphrase string, salt []byte) []byte` using `argon2.IDKey` with constants (e.g., time=3, memory=64*1024, threads=4, keyLen=32).
- Helper `encryptWithKey(plaintext string, key []byte) (string, error)` returning base64(nonce+ciphertext).

**Step 4: Implement `ExportConnectionsByIDsWithPassphrase`**

Add `ExportConnectionsByIDsWithPassphrase(connectionIDs []string, passphrase string) (string, error)`:
- Validate `passphrase != ""`.
- Generate random 16-byte salt.
- Derive key using Argon2id.
- Populate `PasswordEncryption` with mode `passphrase`, salt base64, kdf `argon2id`.
- For each selected connection, encrypt plaintext password with derived key; store as `penc:<base64>`.

**Step 5: Run tests**

Run: `go test ./...`
Expected: FAIL only on pre-existing `main_test.go` signature mismatch.

**Step 6: Commit**

```bash
git add app.go app_export_passphrase_test.go
git commit -m "feat: add passphrase export encryption"
```

---

### Task 2: Passphrase import handling

**Files:**
- Modify: `app.go`
- Test: `app_import_passphrase_test.go`

**Step 1: Write the failing test**

Create `app_import_passphrase_test.go` that:
- Builds export JSON with `password_encryption` mode `passphrase`, salt, and `penc:` values.
- Calls `ImportConnectionsWithPassphrase(json, passphrase)`.
- Verifies imported connection password is available via `GetPassword`.

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL because import method doesn’t exist / password not restored.

**Step 3: Implement import API**

In `app.go`, add:
- `ImportConnectionsWithPassphrase(jsonData, passphrase string) (int, error)`
- `ImportConnectionsFromFileWithPassphrase(filePath, passphrase string) (int, error)`

Behavior:
- Parse `ExportData`.
- If `PasswordEncryption.mode == "passphrase"`, require passphrase; derive key with salt; decrypt each `penc:` value; store into credential store using `SavePassword` (plaintext → machine-bound encryption).
- If mode not passphrase, delegate to existing `ImportConnections`.
- Return clear errors for missing/wrong passphrase (`"passphrase required"`, `"invalid passphrase"`).

**Step 4: Run tests**

Run: `go test ./...` (note existing failure).

**Step 5: Commit**

```bash
git add app.go app_import_passphrase_test.go
git commit -m "feat: import passphrase-encrypted exports"
```

---

### Task 3: Frontend export UI for passphrase

**Files:**
- Modify: `frontend/src/components/AssetList.svelte`
- Modify: `frontend/src/components/ui/InputDialog.svelte` (if needed for confirm)

**Step 1: Write a minimal UI spec (no tests)**

Update export menu to include “导出(跨设备加密)”. When clicked:
- Prompt for passphrase (InputDialog, password type).
- Prompt for confirmation (second InputDialog) and ensure match.
- On success, call `ExportConnectionsByIDsWithPassphrase(ids, passphrase)`.

**Step 2: Implement UI changes**

In `AssetList.svelte`:
- Add new menu item.
- Add `showPassphraseInput`, `passphraseValue` flow.
- Reuse existing selection dialog; only change export handler to use new backend method.

**Step 3: Manual verification**

- Export with passphrase; ensure file saved.
- Cancel/mismatch should show error and abort.

**Step 4: Commit**

```bash
git add frontend/src/components/AssetList.svelte
git commit -m "feat: add passphrase export option"
```

---

### Task 4: Frontend import flow with passphrase

**Files:**
- Modify: `frontend/src/components/AssetList.svelte`

**Step 1: Implement passphrase prompt on import**

Flow:
- On import confirm, attempt `ImportConnectionsFromFileWithPassphrase(filePath, "")`.
- If backend returns “passphrase required”, prompt for passphrase and retry with `ImportConnectionsFromFileWithPassphrase(filePath, passphrase)`.
- If “invalid passphrase”, show error and allow retry.

**Step 2: Manual verification**

- Import passphrase-encrypted file with correct passphrase (succeeds).
- Import with wrong passphrase (error).

**Step 3: Commit**

```bash
git add frontend/src/components/AssetList.svelte
git commit -m "feat: support passphrase import"
```

---

### Task 5: Bindings + docs

**Files:**
- Modify: `frontend/wailsjs/go/main/App.d.ts` (auto-generated)
- Modify: `frontend/wailsjs/go/main/App.js` (auto-generated)
- Modify: `README.md` (optional note)

**Step 1: Regenerate bindings**

Run `wails dev` or `wails build` to regenerate bindings after adding new Go methods.

**Step 2: Document behavior (optional)**

Add a short note to `README.md` about passphrase-based exports being required for cross-device import.

**Step 3: Final verification**

- Run `go test ./...` (note existing failure).
- Manual UI checks for export/import flows.

**Step 4: Commit**

```bash
git add frontend/wailsjs/go/main/App.* README.md
git commit -m "docs: note passphrase export/import"
```
