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
		var initialDiag = data[1] && data[1].stdout ? data[1].stdout : '';

		var brandHeaderHTML =
			'<div style="background:#212121; border-radius:10px; padding:18px 24px; margin-bottom:15px; color:#ffffff; box-shadow:0 4px 15px rgba(0,0,0,0.3); display:flex; align-items:center; gap:20px; border-left:6px solid #00C7E2;">' +
				'<div style="flex-shrink:0;">' +
					'<svg width="50" height="50" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">' +
						'<path d="M50 10 L85 30 L85 70 L50 90 L15 70 L15 30 Z" stroke="url(#clawrt_grad2)" stroke-width="6" fill="none"/>' +
						'<circle cx="50" cy="50" r="15" fill="url(#clawrt_grad2)"/>' +
						'<defs>' +
							'<linearGradient id="clawrt_grad2" x1="0%" y1="0%" x2="100%" y2="0%">' +
								'<stop offset="0%" stop-color="#FF6F00"/>' +
								'<stop offset="33%" stop-color="#C173FF"/>' +
								'<stop offset="66%" stop-color="#3AB5ED"/>' +
								'<stop offset="100%" stop-color="#00C7E2"/>' +
							'</linearGradient>' +
						'</defs>' +
					'</svg>' +
				'</div>' +
				'<div style="flex-grow:1;">' +
					'<div style="font-size:22px; font-weight:800; font-family:Poppins, sans-serif; background:linear-gradient(90deg, #00C7E2, #3AB5ED, #C173FF, #FF6F00); -webkit-background-clip:text; -webkit-text-fill-color:transparent; display:inline-block;">ClawRT AI Copilot</div>' +
					'<span style="font-size:12px; background:rgba(0,199,226,0.15); color:#00C7E2; border:1px solid rgba(0,199,226,0.4); border-radius:4px; padding:2px 8px; margin-left:10px; font-weight:bold;">OpenWrt Agent</span>' +
					'<div style="font-size:13px; color:#CCCCCC; margin-top:3px;">Asistente agéntico local para administración inteligente, auto-sanación y optimización de red.</div>' +
				'</div>' +
			'</div>';

		var chatContainerHTML =
			brandHeaderHTML +
			'<div style="margin-bottom:12px; display:flex; gap:8px; flex-wrap:wrap;">' +
				'<button type="button" class="btn cbi-button cbi-button-action quick-chip" data-prompt="Diagnostica el estado general de la red y ejecuta auto-sanación si es necesario">🩺 Diagnóstico & Auto-Sanación</button>' +
				'<button type="button" class="btn cbi-button cbi-button-apply quick-chip" data-prompt="Audita la seguridad general del router (contraseña root, puertos SSH/LuCI y cortafuegos)">🛡️ Auditoría de Seguridad</button>' +
				'<button type="button" class="btn cbi-button cbi-button-apply quick-chip" data-prompt="Analiza el espectro WiFi y recomiéndame el canal más limpio">✨ Optimizar Canales WiFi</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Detecta clientes WiFi conectados con señal débil (< -80 dBm) que ralentizan la celda">📶 Clientes Pegajosos</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Inspecciona las conexiones conntrack en busca de hosts que saturen la red o posibles escaneos">🛡️ Conntrack & Tráfico</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Lista los clientes conectados en la LAN con fabricante y calidad de señal">📱 Clientes Conectados</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Revisa el estado de Bufferbloat y calidad de servicio SQM">⚡ Bufferbloat / SQM</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral quick-chip" data-prompt="Optimiza la memoria RAM ejecutando el Garbage Collector">🧠 Liberar RAM (GC)</button>' +
			'</div>' +
			'<div id="clawrt-chat-history" style="background:#181818; border:1px solid #333; border-radius:8px; padding:16px; height:420px; overflow-y:auto; font-family:system-ui, -apple-system, sans-serif; display:flex; flex-direction:column; gap:12px; margin-bottom:12px;">' +
				'<div style="align-self:flex-start; max-width:85%; background:#252525; border-left:4px solid #00C7E2; padding:12px 16px; border-radius:6px; color:#f0f0f0; font-size:13px; line-height:1.5;">' +
					'<b>🤖 ClawRT Agent:</b> ¡Hola! Soy tu copiloto autónomo para este router OpenWrt. Puedo asistirte en la configuración de redes de invitados, optimización de canales WiFi, diagnóstico de fallas y mitigación de amenazas en tiempo real.' +
				'</div>' +
			'</div>' +
			'<div style="display:flex; gap:10px;">' +
				'<input type="text" id="clawrt-user-input" class="cbi-input-text" style="flex-grow:1; font-size:14px; padding:10px 14px; border-radius:6px; background:#222; color:#fff; border:1px solid #444;" placeholder="Escribe tu instrucción (ej: Crea una red WiFi de invitados aislada, o diagnostica la conexión WAN)..." />' +
				'<button type="button" id="clawrt-send-btn" class="btn cbi-button cbi-button-apply" style="background:linear-gradient(90deg, #00C7E2, #3AB5ED); color:#fff; font-weight:bold; padding:0 24px;">Enviar</button>' +
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

			function appendMessage(sender, text, isUser) {
				var msgDiv = document.createElement('div');
				msgDiv.style.maxWidth = '85%';
				msgDiv.style.padding = '10px 14px';
				msgDiv.style.borderRadius = '6px';
				msgDiv.style.fontSize = '13px';
				msgDiv.style.lineHeight = '1.5';
				msgDiv.style.wordBreak = 'break-word';

				if (isUser) {
					msgDiv.style.alignSelf = 'flex-end';
					msgDiv.style.background = '#005f73';
					msgDiv.style.color = '#ffffff';
					msgDiv.innerHTML = '<b>👤 Tú:</b> ' + text.replace(/</g, '&lt;').replace(/>/g, '&gt;');
				} else {
					msgDiv.style.alignSelf = 'flex-start';
					msgDiv.style.background = '#252525';
					msgDiv.style.borderLeft = '4px solid #00C7E2';
					msgDiv.style.color = '#f0f0f0';
					
					// Simple Markdown formatting
					var formatted = text
						.replace(/</g, '&lt;')
						.replace(/>/g, '&gt;')
						.replace(/```([\s\S]*?)```/g, '<pre style="background:#111; padding:8px; border-radius:4px; overflow-x:auto; margin:6px 0;">$1</pre>')
						.replace(/`([^`]+)`/g, '<code style="background:#111; padding:2px 5px; border-radius:3px; color:#00C7E2;">$1</code>')
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
				loadingDiv.style.color = '#888';
				loadingDiv.style.fontSize = '12px';
				loadingDiv.style.fontStyle = 'italic';
				loadingDiv.id = 'clawrt-loading-indicator';
				loadingDiv.innerHTML = '⏳ ClawRT está razonando y ejecutando herramientas de diagnóstico en OpenWrt...';
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
