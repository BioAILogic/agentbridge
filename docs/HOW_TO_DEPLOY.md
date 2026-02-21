# SynBridge — How to Deploy an Update

Written for Åsa. No coding knowledge assumed.

---

## What you need open

Three terminal windows:

| Window | What it is | How to open |
|--------|-----------|-------------|
| **WSL** | Linux on your Windows machine. Builds the code and copies files to the server. | Start menu → "Ubuntu" |
| **VPS root** | The actual server in Germany. | See below |
| **Browser** | To verify the result. | Any browser |

---

## How to open the VPS root terminal

In WSL, type:
```
ssh synbridge
```
You are now root on the server. The prompt looks like:
```
[root@localhost ~]#
```

---

## How to deploy an updated landing page (HTML only)

No rebuild needed. Just copy the file.

**Step 1** — In WSL:
```bash
scp -i ~/.ssh/id_ed25519 /mnt/c/Users/asahi/agentbridge/synbridge_landing.html synbridge@87.106.213.239:/tmp/index.html
```

**Step 2** — In VPS root:
```bash
cp /tmp/index.html /opt/synbridge/static/index.html
```

**Step 3** — Refresh https://synbridge.eu in browser. Done.

---

## How to deploy a Go code change (backend update)

Needed when you change any `.go` file.

**Step 1** — Build in WSL:
```bash
bash ~/.lyra/build-and-deploy.sh
```

**Step 2** — Copy binary to server, in WSL:
```bash
scp -i ~/.ssh/id_ed25519 /mnt/c/Users/asahi/agentbridge/bin/synbridge synbridge@87.106.213.239:/tmp/synbridge
```

**Step 3** — Install on server, in VPS root:
```bash
mv /tmp/synbridge /opt/synbridge/bin/synbridge
restorecon -v /opt/synbridge/bin/synbridge
systemctl restart synbridge
systemctl status synbridge
```

The status should say `Active: active (running)`.

**Step 4** — Check health in browser: https://synbridge.eu/health
Should show: `{"db":"ok","status":"ok"}`

---

## How to check if the server is running

In VPS root:
```bash
systemctl status synbridge
```

If it says `active (running)` — all good.
If it says `failed` — run `journalctl -u synbridge -n 50` and send the output to Lyra.

---

## How to push changes to GitHub

In a Windows terminal (PowerShell or CMD), from `C:\Users\asahi\agentbridge`:
```bash
git add -A
git commit -m "describe what you changed"
git push origin main
```

Or let Lyra do it.

---

## If WSL forgets the PATH (commands not found)

Run this at the start of the WSL session:
```bash
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH
```

---

## Key addresses

| Thing | Address |
|-------|---------|
| Live site | https://synbridge.eu |
| Health check | https://synbridge.eu/health |
| Server IP | 87.106.213.239 |
| Server files | `/opt/synbridge/` |
| Binary | `/opt/synbridge/bin/synbridge` |
| Landing page | `/opt/synbridge/static/index.html` |
| Logo | `/opt/synbridge/static/assets/logos/` |
| Database | PostgreSQL 16, database `synbridge`, user `synbridge` |

---

## SSH config (already set up, just for reference)

`~/.ssh/config` on your Windows machine has:
```
Host synbridge        → logs in as root
Host synbridge-app    → logs in as synbridge user
```

So `ssh synbridge` = root on the VPS.
