'use strict';
'require view';
'require fs';
'require ui';

return view.extend({
	load: function() {
		return Promise.all([
			L.resolveDefault(fs.exec('/etc/init.d/clawrt', ['status']), null),
			L.resolveDefault(fs.exec('/usr/bin/clawrt', ['-diagnose']), null),
		]);
	},

	render: function(data) {
		var serviceRunning = data[0] && data[0].code === 0;

		var brandHeaderHTML =
			'<div style="background:#1e1e24; border-radius:12px; padding:20px 24px; margin-bottom:18px; color:#ffffff; box-shadow:0 6px 20px rgba(0,0,0,0.35); border:1px solid #333; border-left:6px solid #00C7E2;">' +
				'<div style="display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:16px;">' +
					'<div style="display:flex; align-items:center; gap:18px;">' +
						'<div style="flex-shrink:0;">' +
							'<svg width="52" height="52" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">' +
								'<path d="M50 10 L85 30 L85 70 L50 90 L15 70 L15 30 Z" stroke="url(#clawrt_grad_copilot)" stroke-width="6" fill="none"/>' +
								'<circle cx="50" cy="50" r="16" fill="url(#clawrt_grad_copilot)"/>' +
								'<defs>' +
									'<linearGradient id="clawrt_grad_copilot" x1="0%" y1="0%" x2="100%" y2="0%">' +
										'<stop offset="0%" stop-color="#FF6F00"/>' +
										'<stop offset="33%" stop-color="#C173FF"/>' +
										'<stop offset="66%" stop-color="#3AB5ED"/>' +
										'<stop offset="100%" stop-color="#00C7E2"/>' +
									'</linearGradient>' +
								'</defs>' +
							'</svg>' +
						'</div>' +
						'<div>' +
							'<div style="font-size:24px; font-weight:800; font-family:Poppins, -apple-system, sans-serif; background:linear-gradient(90deg, #00C7E2, #3AB5ED, #C173FF, #FF6F00); -webkit-background-clip:text; -webkit-text-fill-color:transparent; display:inline-block;">ClawRT AI Copilot</div>' +
							'<span style="font-size:12px; background:rgba(0,199,226,0.18); color:#00C7E2; border:1px solid rgba(0,199,226,0.45); border-radius:4px; padding:2px 8px; margin-left:10px; font-weight:bold;">Autonomous Agent</span>' +
							'<div style="font-size:13px; color:#A0A0A0; margin-top:4px;">Asistente agéntico para diagnóstico autónomo, optimización de espectro, auto-sanación y blindaje.</div>' +
						'</div>' +
					'</div>' +
					'<div style="display:flex; align-items:center; gap:10px;">' +
						'<div style="background:#282830; border:1px solid #444; border-radius:8px; padding:8px 14px; text-align:center;">' +
							'<div style="font-size:11px; color:#888; text-transform:uppercase; font-weight:bold;">Estado Daemon</div>' +
							'<div style="font-size:13px; font-weight:bold; color:' + (serviceRunning ? '#00e676' : '#ff5252') + ';">' + (serviceRunning ? '● Activo (Watchdog & UBUS)' : '○ Inactivo') + '</div>' +
						'</div>' +
					'</div>' +
				'</div>' +

				// Live Cascading Connectivity Bar
				'<div style="margin-top:16px; padding-top:14px; border-top:1px solid #333; display:flex; align-items:center; gap:8px; flex-wrap:wrap; font-size:12px;">' +
					'<span style="color:#888; font-weight:bold;">ENLACE:</span>' +
					'<span style="background:#162a2c; color:#00C7E2; border:1px solid #00C7E2; border-radius:4px; padding:3px 8px;">Gateway 🟢</span>' +
					'<span style="color:#555;">➔</span>' +
					'<span style="background:#162a2c; color:#00C7E2; border:1px solid #00C7E2; border-radius:4px; padding:3px 8px;">DNS Local 🟢</span>' +
					'<span style="color:#555;">➔</span>' +
					'<span style="background:#162a2c; color:#00C7E2; border:1px solid #00C7E2; border-radius:4px; padding:3px 8px;">DNS Público 🟢</span>' +
					'<span style="color:#555;">➔</span>' +
					'<span style="background:#162a2c; color:#00C7E2; border:1px solid #00C7E2; border-radius:4px; padding:3px 8px;">WAN Internet 🟢</span>' +
				'</div>' +
			'</div>';

		var categoryTabsHTML =
			'<div style="margin-bottom:12px;">' +
				'<div style="display:flex; gap:6px; margin-bottom:10px; border-bottom:1px solid #333; padding-bottom:8px;">' +
					'<button type="button" class="tab-btn active-tab" data-tab="tab-diag" style="background:#282830; color:#00C7E2; border:1px solid #00C7E2; border-radius:6px; padding:6px 12px; font-weight:bold; cursor:pointer; font-size:12px;">🩺 Diagnóstico & Rescate</button>' +
					'<button type="button" class="tab-btn" data-tab="tab-wifi" style="background:#1e1e24; color:#aaa; border:1px solid #444; border-radius:6px; padding:6px 12px; font-weight:bold; cursor:pointer; font-size:12px;">📶 WiFi & Clientes</button>' +
					'<button type="button" class="tab-btn" data-tab="tab-sec" style="background:#1e1e24; color:#aaa; border:1px solid #444; border-radius:6px; padding:6px 12px; font-weight:bold; cursor:pointer; font-size:12px;">🛡️ Seguridad & VPN</button>' +
					'<button type="button" class="tab-btn" data-tab="tab-perf" style="background:#1e1e24; color:#aaa; border:1px solid #444; border-radius:6px; padding:6px 12px; font-weight:bold; cursor:pointer; font-size:12px;">⚡ Rendimiento & Flash</button>' +
				'</div>' +

				// Tab Contents
				'<div id="tab-diag" class="tab-content" style="display:flex; gap:8px; flex-wrap:wrap;">' +
					'<button type="button" class="btn cbi-button cbi-button-action quick-chip" data-prompt="Diagnostica la conectividad y ejecuta auto-sanación en caso de anomalías">🩺 Diagnóstico & Auto-Sanación</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Inspecciona el estado de múltiples enlaces WAN (mwan3) y balanceo de carga">🔀 Multi-WAN & Failover</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Muestra el estado general del sistema, procesador, memoria RAM y uptime">📊 Estado General</button>' +
				'</div>' +

				'<div id="tab-wifi" class="tab-content" style="display:none; gap:8px; flex-wrap:wrap;">' +
					'<button type="button" class="btn cbi-button cbi-button-apply quick-chip" data-prompt="Analiza el espectro WiFi de redes vecinas y calcula el canal óptimo">✨ Optimizar Canales WiFi</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Detecta clientes WiFi conectados con señal débil (< -80 dBm) que ralentizan la red">📶 Clientes Pegajosos</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Lista los dispositivos conectados en la LAN con fabricante y dirección IP">📱 Clientes Conectados</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Genera el código QR para conectarse a la red WiFi del router">📷 Código QR WiFi</button>' +
				'</div>' +

				'<div id="tab-sec" class="tab-content" style="display:none; gap:8px; flex-wrap:wrap;">' +
					'<button type="button" class="btn cbi-button cbi-button-apply quick-chip" data-prompt="Audita la seguridad general del router (contraseña root, puertos SSH/LuCI expuestos y firewall)">🛡️ Auditoría de Seguridad</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Inspecciona el estado de túneles WireGuard y reconecta peers caídos">🔒 Túneles WireGuard</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Audita la privacidad DNS (DoH/DoT cifrado) y estado del bloqueador AdBlock">🌐 Privacidad DNS & AdBlock</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Inspecciona las conexiones de la tabla conntrack en busca de saturación o escaneos">🛡️ Conntrack & Tráfico</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Escanea los 9 puertos críticos en los dispositivos de la red local">🛡️ Escáner de Puertos LAN</button>' +
				'</div>' +

				'<div id="tab-perf" class="tab-content" style="display:none; gap:8px; flex-wrap:wrap;">' +
					'<button type="button" class="btn cbi-button cbi-button-action quick-chip" data-prompt="Audita el espacio libre en Flash (/overlay) y lista paquetes con actualizaciones">💾 Espacio en Flash & Paquetes</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Revisa el estado de Bufferbloat y calidad de servicio SQM (Cake)">⚡ Bufferbloat / SQM</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Optimiza la memoria RAM ejecutando el Garbage Collector">🧠 Liberar RAM (GC)</button>' +
					'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Genera un respaldo cifrado de la configuración de OpenWrt (AES-GCM)">🔐 Respaldo Cifrado</button>' +
				'</div>' +
			'</div>';

		var chatContainerHTML =
			brandHeaderHTML +
			categoryTabsHTML +
			'<div id="clawrt-chat-history" style="background:#141418; border:1px solid #2d2d35; border-radius:10px; padding:18px; height:440px; overflow-y:auto; font-family:system-ui, -apple-system, sans-serif; display:flex; flex-direction:column; gap:14px; margin-bottom:14px; box-shadow:inset 0 2px 8px rgba(0,0,0,0.5);">' +
				'<div style="align-self:flex-start; max-width:85%; background:#202028; border-left:4px solid #00C7E2; padding:14px 18px; border-radius:8px; color:#f0f0f0; font-size:13px; line-height:1.6; box-shadow:0 2px 6px rgba(0,0,0,0.2);">' +
					'<b>🤖 ClawRT Agent:</b> ¡Hola! Soy tu copiloto autónomo para este router OpenWrt. Puedo asistirte en la configuración declarativa de redes de invitados, optimización de canales WiFi, diagnóstico de fallas, túneles WireGuard y mitigación de amenazas.' +
				'</div>' +
			'</div>' +
			'<div style="display:flex; gap:10px;">' +
				'<input type="text" id="clawrt-user-input" class="cbi-input-text" style="flex-grow:1; font-size:14px; padding:12px 16px; border-radius:8px; background:#1e1e24; color:#fff; border:1px solid #3d3d48; outline:none;" placeholder="Escribe tu instrucción (ej: Audita la seguridad del router, optimiza el canal WiFi o diagnostica la red)..." />' +
				'<button type="button" id="clawrt-send-btn" class="btn cbi-button cbi-button-apply" style="background:linear-gradient(90deg, #00C7E2, #3AB5ED); color:#fff; font-weight:bold; padding:0 26px; border-radius:8px; border:none; cursor:pointer; font-size:14px;">Enviar</button>' +
			'</div>';

		var viewNode = E('div', { class: 'cbi-map' }, [
			E('div', { class: 'cbi-section' }, [
				E('div', {}, [])
			])
		]);
		viewNode.innerHTML = chatContainerHTML;

		// Interactive handlers
		setTimeout(function() {
			var historyEl = document.getElementById('clawrt-chat-history');
			var inputEl = document.getElementById('clawrt-user-input');
			var sendBtn = document.getElementById('clawrt-send-btn');

			// Tab switching logic
			var tabButtons = viewNode.querySelectorAll('.tab-btn');
			var tabContents = viewNode.querySelectorAll('.tab-content');

			tabButtons.forEach(function(btn) {
				btn.addEventListener('click', function() {
					tabButtons.forEach(function(b) {
						b.style.background = '#1e1e24';
						b.style.color = '#aaa';
						b.style.borderColor = '#444';
						b.classList.remove('active-tab');
					});
					tabContents.forEach(function(c) {
						c.style.display = 'none';
					});

					btn.style.background = '#282830';
					btn.style.color = '#00C7E2';
					btn.style.borderColor = '#00C7E2';
					btn.classList.add('active-tab');

					var targetId = btn.getAttribute('data-tab');
					var targetContent = document.getElementById(targetId);
					if (targetContent) {
						targetContent.style.display = 'flex';
					}
				});
			});

			function appendMessage(sender, text, isUser) {
				var msgDiv = document.createElement('div');
				msgDiv.style.maxWidth = '85%';
				msgDiv.style.padding = '12px 16px';
				msgDiv.style.borderRadius = '8px';
				msgDiv.style.fontSize = '13px';
				msgDiv.style.lineHeight = '1.6';
				msgDiv.style.wordBreak = 'break-word';
				msgDiv.style.boxShadow = '0 2px 6px rgba(0,0,0,0.2)';

				if (isUser) {
					msgDiv.style.alignSelf = 'flex-end';
					msgDiv.style.background = '#005f73';
					msgDiv.style.color = '#ffffff';
					msgDiv.innerHTML = '<b>👤 Tú:</b> ' + text.replace(/</g, '&lt;').replace(/>/g, '&gt;');
				} else {
					msgDiv.style.alignSelf = 'flex-start';
					msgDiv.style.background = '#202028';
					msgDiv.style.borderLeft = '4px solid #00C7E2';
					msgDiv.style.color = '#f0f0f0';
					
					// Formatted Markdown renderer
					var formatted = text
						.replace(/</g, '&lt;')
						.replace(/>/g, '&gt;')
						.replace(/```([\s\S]*?)```/g, '<pre style="background:#111116; padding:10px; border-radius:6px; overflow-x:auto; margin:8px 0; border:1px solid #333; font-family:monospace; color:#3AB5ED;">$1</pre>')
						.replace(/`([^`]+)`/g, '<code style="background:#111116; padding:2px 6px; border-radius:4px; color:#00C7E2; font-family:monospace;">$1</code>')
						.replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>')
						.replace(/\*([^*]+)\*/g, '<i>$1</i>')
						.replace(/\n/g, '<br/>');

					msgDiv.innerHTML = '<b>🤖 ClawRT Agent:</b> ' + formatted;
				}

				historyEl.appendChild(msgDiv);
				historyEl.scrollTop = historyEl.scrollHeight;
			}

			function sendPrompt(promptText) {
				if (!promptText || !promptText.trim()) return;
				appendMessage('Tú', promptText, true);
				inputEl.value = '';
				inputEl.disabled = true;
				sendBtn.disabled = true;

				var loadingDiv = document.createElement('div');
				loadingDiv.style.alignSelf = 'flex-start';
				loadingDiv.style.color = '#00C7E2';
				loadingDiv.style.fontSize = '12px';
				loadingDiv.style.fontStyle = 'italic';
				loadingDiv.id = 'clawrt-loading-indicator';
				loadingDiv.innerHTML = '⏳ <i>ClawRT está razonando y ejecutando diagnósticos en OpenWrt...</i>';
				historyEl.appendChild(loadingDiv);
				historyEl.scrollTop = historyEl.scrollHeight;

				fs.exec('/usr/bin/clawrt', ['-query', promptText]).then(function(res) {
					var loadInd = document.getElementById('clawrt-loading-indicator');
					if (loadInd) loadInd.remove();

					var out = res && res.stdout ? res.stdout.trim() : (res && res.stderr ? res.stderr.trim() : 'Sin respuesta del agente.');
					appendMessage('ClawRT Agent', out, false);

					inputEl.disabled = false;
					sendBtn.disabled = false;
					inputEl.focus();
				}).catch(function(err) {
					var loadInd = document.getElementById('clawrt-loading-indicator');
					if (loadInd) loadInd.remove();
					appendMessage('ClawRT Agent', '❌ Error al ejecutar consulta: ' + err, false);
					inputEl.disabled = false;
					sendBtn.disabled = false;
				});
			}

			if (sendBtn) {
				sendBtn.addEventListener('click', function() {
					sendPrompt(inputEl.value);
				});
			}

			if (inputEl) {
				inputEl.addEventListener('keypress', function(e) {
					if (e.key === 'Enter') {
						sendPrompt(inputEl.value);
					}
				});
			}

			var chips = viewNode.querySelectorAll('.quick-chip');
			chips.forEach(function(chip) {
				chip.addEventListener('click', function() {
					var p = chip.getAttribute('data-prompt');
					if (p) sendPrompt(p);
				});
			});

		}, 200);

		return viewNode;
	},

	handleSave: null,
	handleSaveApply: null,
	handleReset: null
});
