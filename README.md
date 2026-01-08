# KSeF QR Code II Generator

Generate QR Code II for KSeF invoices visualization in minutes.

Are you wondering whether it is possible to use a single, static QR Code II and place it on all invoice visualizations 👀?

Of course it is .

During verification, KSeF does not actually validate what has been signed. If your certificate is valid and you have the authority to issue invoices on behalf of the specified NIP, the QR code will pass verification successfully.

How can this be proven 🧭 ?

Sign any file from your computer using a valid KSeF certificate — or sign nothing at all (in which case the application will sign the default text: “KSeF is dead baby, KSeF is dead…”). Then verify the QR Code by scanning it with your phone or by opening the link generated and displayed in the console.

You need valid KSeF certificates for a given environment ⚠️

## Motivation

This project was created as a pragmatic response to reality — not theory.

According to the official KSeF documentation, the QR Code II verification mechanism is supposed to be based on a cryptographic signature calculated over the hash of the invoice XML on which the QR code is placed. In theory, this should ensure integrity and bind the QR code to the actual content of the invoice.

In practice, however, KSeF verifies this signature rather loosely 🚧.

The uncomfortable fact is that KSeF currently accepts a QR Code II as fully valid even if the signature was calculated over any arbitrary SHA-256 value — not necessarily the hash of the invoice XML itself. As long as the signature structure looks correct, the system is satisfied.

This tool does not exist to encourage bypassing requirements or weakening security on purpose. Quite the opposite.

The motivation behind it is to help those who:

- are still in the process of designing a proper, secure storage for KSeF certificates,
- have not yet implemented hardware-backed key protection, HSMs, or dedicated key vaults,
- or simply need an emergency fallback to keep their invoicing process operational.

With this tool, you can generate a single, valid QR Code that can be placed on all invoice visualizations issued in offline mode or during a KSeF outage 🖨️.

From a security perspective, storing a private key used for signing is not trivial. It requires time, architecture decisions, threat modeling, and a careful implementation. Pretending otherwise would be irresponsible.

Since the KSeF platform itself currently does not strictly enforce what it formally requires, this tool provides a controlled and transparent workaround — clearly exposing the gap, not hiding it.

Think of it as a temporary life raft, not a long-term security strategy.

## Usage

````shell
qr-gen -c sigining_cert.crt -k sigining_cert.key -n 1111111111
````

or with real data:

````shell
qr-gen -c sigining_cert.crt -k sigining_cert.key -n 1111111111 --in invoice.xml
````

If you set `KSEF_KEY_PASSWORD` environment variable, you will be not prompted for the password.

You can download binaries for Windows and Linux on the releases page on GitHub.

Happy KSeFing!