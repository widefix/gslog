# macOS Code Signing and Notarization

Signed and notarized macOS binaries pass Gatekeeper checks without prompting users. Without signing, users may see *"Apple could not verify that git-squash-tree is free of malware"* and must run `xattr -d com.apple.quarantine` or allow the app in System Settings.

**Cost:** Apple Developer Program ($99/year). No additional fees for signing or notarization.

---

## Prerequisites

1. **Apple Developer account** — [developer.apple.com](https://developer.apple.com)
2. **Xcode Command Line Tools** — `xcode-select --install` if needed
3. **macOS** — Signing and notarization must run on a Mac

---

## Setup (one-time)

### 1. Create a Developer ID Application certificate

1. Open [Certificates, Identifiers & Profiles](https://developer.apple.com/account/resources/certificates/list).
2. Click **+** to add a certificate.
3. Choose **Developer ID Application**.
4. Follow the prompts; use the private key stored in your Mac’s Keychain.

### 2. Store notarization credentials

You need an **app-specific password** for your Apple ID (not your normal password):

1. Go to [appleid.apple.com](https://appleid.apple.com) → Sign-In and Security → App-Specific Passwords.
2. Generate a new app-specific password (e.g. name: `squash-tree notary`).
3. In Terminal:

```bash
xcrun notarytool store-credentials "AC_PASSWORD" \
  --apple-id "jasirguzman@gmail.com" \
  --team-id "5MYVLY5D2P" \
  --password "lohz-cyjl-bgjo-pulp"
```

- Replace `AC_PASSWORD` with a profile name (this is just a label).
- Use the app-specific password you created.
- Find your Team ID in [Apple Developer → Membership](https://developer.apple.com/account/#MembershipDetailsCard).

### 3. Get your signing identity

List installed Developer ID certificates:

```bash
security find-identity -v -p codesigning | grep "Developer ID Application"
```

You’ll see something like:

```
1) ABCD1234... "Developer ID Application: Your Name (TEAM_ID)"
```

Use that full string as `SIGNING_IDENTITY`.

---

## Build with signing

Set the environment variables and run the build:

```bash
export SIGNING_IDENTITY="Developer ID Application: Your Name (TEAM_ID)"
export NOTARY_KEYCHAIN_PROFILE="AC_PASSWORD"

./scripts/build-release.sh v0.1.0
```

The script will:

1. Build macOS binaries for Intel and Apple Silicon.
2. Sign each with your Developer ID certificate.
3. Submit each to Apple for notarization (as a zip; Apple notarizes the signed binary inside).
4. Package them into the release archives.

Stapling (embedding the notarization ticket in the file) is only supported for `.app`, `.dmg`, and `.pkg` — not raw CLI binaries. Gatekeeper still verifies notarization **online** using the binary’s signature, so users won’t see a warning.

Notarization usually takes 1–3 minutes. If either variable is unset, macOS binaries are built but left unsigned.

---

## Troubleshooting

**"The signature of the binary is invalid"** — Ensure `SIGNING_IDENTITY` matches the full certificate name from `security find-identity`.

**Notarization fails** — Check the log:

```bash
xcrun notarytool log SUBMISSION_ID --keychain-profile "AC_PASSWORD"
```
