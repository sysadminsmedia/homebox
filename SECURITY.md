# Security Policy

## Supported Versions

Homebox is maintained by [Sysadmins Media](https://github.com/sysadminsmedia), who took over the
project at **v0.11.0** (June 2024). Everything released before that came from the original upstream
project and is not maintained here.

Security reports are accepted for:

| Version | Supported |
| ------- | --------- |
| The latest release | Yes |
| The `main` branch | Yes |
| Any other older release | No |
| Anything before `v0.11.0` | No — not our code |

**We only issue security advisories, CVEs, and reporter credit for issues that are present on the
latest release or on `main` at the time of the report.** Nothing else is eligible.

In practice this means:

- **Check `main` before you report.** A large number of reports we receive describe bugs that were
  already fixed, sometimes years earlier. If the code you are looking at does not exist on `main`
  anymore, there is nothing for us to fix and nothing to credit.
- **Reports against pre-`v0.11.0` versions are closed without assessment.** That code predates our
  stewardship. This includes `v0.10.x` and earlier. We will not issue an advisory or credit for
  them under any circumstances, even if the same bug once existed here.
- **Reports against an old-but-supported-era release are only eligible if the issue still reproduces
  on the latest release or `main`.** "I found this in v0.20.0" is not by itself a valid report if
  the current code is fine. Please state which version or commit you tested.
- **Running an outdated version is not a vulnerability.** If the fix is to upgrade, upgrade.

We are happy to receive good reports and we do credit researchers who follow this. We are not able
to spend maintainer time re-triaging historical code, and automated scans of old container images
are not useful to us.

## Reporting a Vulnerability

Please open a normal public issue for minor security issues or general security inquiries.

For major or critical security issues, please
[open a private GitHub security advisory](https://github.com/sysadminsmedia/homebox/security/advisories/new).

Please include:

- The **version or commit hash** you tested against, and confirmation that you checked it still
  reproduces on the latest release or `main`.
- Steps to reproduce, ideally as concrete requests or a short script.
- What an attacker gains, and any preconditions required (e.g. multiple groups, a reverse proxy, a
  cache in front of the app).

We do not take reports at face value, however they were produced. If you used an AI assistant or an
automated scanner, you are accountable for every claim in the report you send: reproduce the issue
yourself against the latest release or `main`, confirm the code you describe still exists there, and
strip anything you have not personally verified. Reports that read as unreviewed tool output —
invented file paths, line numbers, or function names, findings already fixed upstream, or generic
scanner output with no working proof of concept — are closed without assessment. Repeated
submissions of that kind will be treated as spam.
