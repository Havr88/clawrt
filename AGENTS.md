# 🤖 AGENTS.md — ClawRT Architecture & Fine-Tuning Specification

This document defines the agentic execution engine, autonomous loops, tool invocation boundaries, context pruning algorithms, and hardware adaptation strategies for **ClawRT**.

---

## ⚙️ 1. Hardware Adaptation Engine (Tier Scaling)

ClawRT dynamically adapts its memory consumption, prompt context size, and background worker routines based on the detected hardware specs of the OpenWrt router:

| Hardware Tier | OpenWrt Target Specs | System Prompt | Max Iterations | Auto-Sleep Timeout | Supported Features & Capabilities |
|:---|:---|:---|:---:|:---:|:---|
| **TierMinimal** | **Flash:** $\ge 8\text{ MB}$ *(Bare Minimum)*<br>**RAM:** $64\text{ MB}$ *(Xiaomi Mi 4C)* | ~1 KB (Ultra-compact) | 5 steps | **5 minutes** | FastPath L1/L2, Telegram polling, LuCI Copilot, basic DHCP/WiFi, Self-Healing Watchdog FSM |
| **TierMedium** | **Flash:** $\ge 16\text{ MB}$ *(Recommended)*<br>**RAM:** $128\text{ MB}$ *(Standard Routers)* | ~2 KB (Standard) | 8 steps | **5 minutes** | Full port scanning, Conntrack Guard, WiFi Optimizer, SQM QoS, UBUS Listener, Supabase Realtime |
| **TierFull** | **Flash:** $\ge 32\text{ MB}$ / NVM<br>**RAM:** $> 256\text{ MB}$ *(Linksys Velop/x86)* | ~4 KB (Extended) | 12 steps | 10 minutes | Declarative Intent Engine, deep packet inspection, pgvector RAG, Cloudflare D1/R2, Local SLM fallback |

---

## 🧠 2. Autonomous Background Loops & Memory Optimization

To maintain a minimal RAM footprint (~2–3 MB in idle) while guaranteeing autonomous operations:
1. **Self-Healing Watchdog Loop**:
   - Executes periodic and event-triggered connectivity health checks (Gateway ➔ DNS Local ➔ Public DNS ➔ Internet Ping).
   - Automatically executes staged self-healing when anomalies are detected (dnsmasq reload, WAN renegotiation `ifup wan`, firewall restart).
2. **UBUS Reactive Listener**:
   - Subscribes to `/var/run/ubus/ubus.sock` stream for `network.interface`, `hostapd`, and `procd` events with zero subshell overhead.
3. **Idle Monitor Worker & Auto-Sleep (5 Minutes)**:
   - When idle for **5 minutes**, ClawRT calls Go `runtime.GC()` followed by `debug.FreeOSMemory()`.
   - Unused heap memory is returned to the OpenWrt kernel (`musl/glibc`).
4. **Instant Wakeup**: Upon receiving a Telegram message, LuCI Copilot request, or `/etc/hotplug.d/` event, the agent wakes up instantly with zero cold-start delay.

---

## 🛠️ 3. Skill & Tool Execution Architecture

All tools must adhere to strict input sanitization, 15-second execution timeouts, and output truncation ($\le 4\text{ KB}$):

1. `get_system_info`: Inspect CPU load, uptime, RAM, and OpenWrt release.
2. `get_network_status`: Inspect WAN/LAN interfaces, IP addresses, and traffic stats.
3. `get_dhcp_leases`: Read `/tmp/dhcp.leases` + UCI static leases, resolve vendor OUI, detect randomized MACs.
4. `get_wifi_qr`: Generate ASCII QR code payload for instant WiFi connection.
5. `scan_lan_ports`: Perform lightweight TCP connect scan on 9 critical ports.
6. `read_uci_config`: Read UCI configuration sections (`/etc/config/*`).
7. `write_uci_config`: Execute typed UCI setter with snapshot backup, `fw4 check`, and automatic rollback.
8. `exec_safe_cmd`: Execute safe diagnostic commands (ping, traceroute, nslookup, logread, df, free).
9. `restart_service`: Manage safe system services (`/etc/init.d/*`).
10. `self_healing_diagnostic`: Execute comprehensive connectivity diagnosis and staged hot-repair.
11. `analyze_conntrack_traffic`: Inspect `/proc/net/nf_conntrack` for top bandwidth hogs, port scans, and SYN flood threats.
12. `block_abuser_ip`: Isolate malicious or abusive hosts in firewall (nftables / fw4).
13. `optimize_wifi_channels`: Scan wireless spectrum and calculate least congested channels in 2.4 GHz & 5 GHz.
14. `manage_sqm_qos`: Inspect and configure SQM (Cake / FQ_Codel) to eliminate Bufferbloat.
15. `execute_intent_plan`: Execute declarative multi-step plans (guest WiFi isolation, port forwarding, parental restrictions) with atomic verification.
16. `backup_to_supabase`: Generate full OpenWrt sysupgrade backup for offsite storage.
