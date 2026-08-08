'use strict';
'require view';
'require form';
'require fs';
'require ui';

var providerDefaults = {
	'bynara': {
		url: 'https://router.bynara.id/v1',
		model: 'gpt-4o-mini',
		fallback: 'deepseek-chat'
	},
	'groq': {
		url: 'https://api.groq.com/openai/v1',
		model: 'llama-3.3-70b-versatile',
		fallback: 'llama-3.1-8b-instant'
	},
	'openrouter': {
		url: 'https://openrouter.ai/api/v1',
		model: 'meta-llama/llama-3.3-70b-instruct:free',
		fallback: 'google/gemini-2.0-flash-exp:free'
	},
	'deepseek': {
		url: 'https://api.deepseek.com/v1',
		model: 'deepseek-chat',
		fallback: 'deepseek-reasoner'
	},
	'openai': {
		url: 'https://api.openai.com/v1',
		model: 'gpt-4o-mini',
		fallback: 'gpt-4o'
	},
	'gemini': {
		url: 'https://generativelanguage.googleapis.com/v1beta/openai/',
		model: 'gemini-1.5-flash',
		fallback: 'gemini-2.0-flash-exp'
	},
	'mistral': {
		url: 'https://api.mistral.ai/v1',
		model: 'mistral-small-latest',
		fallback: 'open-mistral-7b'
	},
	'ollama': {
		url: 'http://192.168.1.100:11434/v1',
		model: 'llama3.2',
		fallback: 'qwen2.5-coder'
	},
	'custom': {
		url: 'https://...',
		model: 'custom-model',
		fallback: 'fallback-model'
	}
};

return view.extend({
	load: function() {
		return Promise.all([
			L.resolveDefault(fs.exec('/etc/init.d/clawrt', ['status']), null),
			L.resolveDefault(fs.exec('/bin/ps', ['w']), null),
			L.resolveDefault(fs.exec('/sbin/logread', ['-e', 'clawrt', '-l', '25']), null),
		]);
	},

	render: function(data) {
		var m, s, o;

		var serviceRunning = data[0] && data[0].code === 0;
		var psOutput = data[1] && data[1].stdout ? data[1].stdout : '';
		var logsOutput = data[2] && data[2].stdout ? data[2].stdout : 'No hay registros disponibles de clawrt.';

		var ramUsageStr = 'N/A';
		if (serviceRunning) {
			var lines = psOutput.split('\n');
			for (var i = 0; i < lines.length; i++) {
				if (lines[i].indexOf('/usr/bin/clawrt') !== -1) {
					var parts = lines[i].trim().split(/\s+/);
					if (parts.length >= 5) {
						ramUsageStr = parts[4] + ' (Virtual Memory)';
					}
					break;
				}
			}
		}

		m = new form.Map('clawrt',
			_('ClawRT AI — Agente Agéntico para OpenWrt 🐾⚡'),
			_('Panel de Administración y Control del Agente de Inteligencia Artificial para tu Router OpenWrt.'));

		/* ── 1. Estado del Servicio y Métricas en Tiempo Real ── */
		s = m.section(form.NamedSection, '_status', 'clawrt', _('Panel de Estado & Botones de Acción Rápida'));
		s.anonymous = true;
		s.addremove = false;

		o = s.option(form.DummyValue, '_svc_status', _('Estado de Ejecución del Agente'));
		o.cfgvalue = function() {
			return serviceRunning
				? '<span style="color:#2ecc71;font-weight:bold">● ClawRT Activo (Running)</span> &nbsp;|&nbsp; 🧠 <b>RAM Agente:</b> ' + ramUsageStr
				: '<span style="color:#e74c3c;font-weight:bold">● ClawRT Detenido / En Espera</span>';
		};
		o.rawhtml = true;

		o = s.option(form.DummyValue, '_actions', _('Pruebas Diagnósticas Rápidas'));
		o.cfgvalue = function() {
			return '<div style="margin-top:5px; display:flex; gap:10px; flex-wrap:wrap;">' +
				'<button type="button" class="btn cbi-button cbi-button-apply" id="btn_test_llm">🧪 Probar Conexión LLM</button>' +
				'<button type="button" class="btn cbi-button cbi-button-action" id="btn_fetch_models">🔍 Detectar Modelos en Vivo</button>' +
				'<button type="button" class="btn cbi-button cbi-button-neutral" id="btn_test_telegram">💬 Probar Bot Telegram</button>' +
				'<button type="button" class="btn cbi-button cbi-button-remove" id="btn_clear_facts">🧹 Vaciar Hechos Aprendidos</button>' +
				'</div>';
		};
		o.rawhtml = true;

		/* ── 2. Configuración General e Idioma (i18n) ── */
		s = m.section(form.TypedSection, 'core', _('General & Idioma de Respuesta (i18n)'));
		s.anonymous = true;
		s.addremove = false;

		o = s.option(form.Flag, 'enabled', _('Habilitar Servicio ClawRT'),
			_('Activa o desactiva la ejecución en segundo plano de ClawRT al iniciar el router.'));
		o.default = '1';
		o.rmempty = false;

		o = s.option(form.ListValue, 'language', _('Idioma Preferido (Response Language)'),
			_('Idioma en el que responde el agente por Telegram y en mensajes de sistema.'));
		o.value('auto', _('Auto-detectar (Idioma del Usuario en Telegram / Fallback)'));
		o.value('es', _('Español (Spanish)'));
		o.value('en', _('English'));
		o.value('fr', _('Français (French)'));
		o.value('pt', _('Português (Portuguese)'));
		o.value('it', _('Italiano (Italian)'));
		o.value('ru', _('Русский (Russian)'));
		o.value('zh', _('中文 (Chinese)'));
		o.value('ja', _('日本語 (Japanese)'));
		o.value('ar', _('العربية (Arabic)'));
		o.default = 'auto';

		/* ── 3. Configuración de Telegram (CGNAT Safe) ── */
		s = m.section(form.TypedSection, 'telegram', _('Canal de Telegram (CGNAT Safe)'));
		s.anonymous = true;
		s.addremove = false;
		s.description = _('Configura el Bot Token otorgado por @BotFather y tus Chat IDs autorizados.');

		o = s.option(form.Value, 'bot_token', _('Token del Bot de Telegram'),
			_('Formato: 123456789:ABCdef... (obtenido de @BotFather)'));
		o.password = true;
		o.placeholder = '123456789:ABCdefGHIjklMNO...';
		o.rmempty = false;

		o = s.option(form.DynamicList, 'chat_id', _('Chat IDs Autorizados'),
			_('Solo los Chat IDs listados aquí podrán enviar comandos al router.'));
		o.placeholder = '987654321';
		o.datatype = 'integer';

		/* ── 4. Base de Datos Externa Optativa (Memoria a Largo Plazo) ── */
		s = m.section(form.TypedSection, 'db', _('Base de Datos Externa Optativa (Memoria a Largo Plazo)'));
		s.anonymous = true;
		s.addremove = false;
		s.description = _('Almacena historial de eventos y memoria agéntica fuera del router sin consumir los 64MB de RAM.');

		var dbProv = s.option(form.ListValue, 'provider', _('Proveedor de Base de Datos Externa'));
		dbProv.value('none', _('Ninguno (Almacenamiento Local en /tmp) [Default]'));
		dbProv.value('redis', _('Upstash Redis (Serverless HTTP Redis)'));
		dbProv.value('d1', _('Cloudflare D1 (Serverless SQLite API)'));
		dbProv.value('supabase', _('Supabase (PostgreSQL REST API)'));
		dbProv.default = 'none';

		var dbUrl = s.option(form.Value, 'url', _('URL del Endpoint de Base de Datos'));
		dbUrl.depends('provider', 'redis');
		dbUrl.depends('provider', 'd1');
		dbUrl.depends('provider', 'supabase');
		dbUrl.placeholder = 'https://...upstash.io o https://...supabase.co';
		dbUrl.rmempty = true;

		var dbToken = s.option(form.Value, 'token', _('Token de Autenticación / Bearer Token'));
		dbToken.password = true;
		dbToken.depends('provider', 'redis');
		dbToken.depends('provider', 'd1');
		dbToken.depends('provider', 'supabase');
		dbToken.placeholder = 'Bearer eyJhbGciOi...';
		dbToken.rmempty = true;

		/* ── 5. Selector de Proveedor LLM (Model Registry & Routing Engine) ── */
		s = m.section(form.TypedSection, 'llm', _('Selector de Proveedor LLM (Model Registry & Enrutamiento)'));
		s.anonymous = true;
		s.addremove = false;
		s.description = _('Selecciona tu proveedor preferido comercial o gratuito con URL y modelos predeterminados automáticos.');

		var llmProv = s.option(form.ListValue, 'provider', _('Proveedor de Inteligencia Artificial'));
		llmProv.value('bynara', _('NaraRouter Gateway (https://router.bynara.id/v1) [Default]'));
		llmProv.value('groq', _('Groq Cloud (https://api.groq.com/openai/v1) [Tier Gratuito]'));
		llmProv.value('openrouter', _('OpenRouter (https://openrouter.ai/api/v1) [Modelos Gratis]'));
		llmProv.value('deepseek', _('DeepSeek API (https://api.deepseek.com/v1) [V3 & R1]'));
		llmProv.value('openai', _('OpenAI Directo (https://api.openai.com/v1)'));
		llmProv.value('gemini', _('Google Gemini (https://generativelanguage.googleapis.com/.../openai/)'));
		llmProv.value('mistral', _('Mistral AI (https://api.mistral.ai/v1)'));
		llmProv.value('ollama', _('Ollama Local (http://192.168.1.100:11434/v1) [Servidor LAN]'));
		llmProv.value('custom', _('Personalizado / Custom API Endpoint'));
		llmProv.default = 'bynara';

		var baseUrlOpt = s.option(form.Value, 'base_url', _('URL Base de la API LLM'),
			_('URL prediseñada del proveedor seleccionado (se auto-completa al cambiar de proveedor).'));
		baseUrlOpt.placeholder = 'https://router.bynara.id/v1';
		baseUrlOpt.default = 'https://router.bynara.id/v1';

		var apiKeyOpt = s.option(form.Value, 'api_key', _('API Key / Clave de Acceso'),
			_('Clave de API otorgada por tu proveedor elegido (opcional para Ollama local).'));
		apiKeyOpt.password = true;
		apiKeyOpt.placeholder = 'sk-...';

		var modelOpt = s.option(form.Value, 'model', _('Modelo Principal'),
			_('Nombre del modelo principal recomendado para el proveedor seleccionado.'));
		modelOpt.placeholder = 'gpt-4o-mini';
		modelOpt.default = 'gpt-4o-mini';

		var fallbackOpt = s.option(form.Value, 'fallback_model', _('Modelo Respaldo (Fallback Model)'),
			_('Modelo secundario de respaldo si el principal falla por cuota o timeout.'));
		fallbackOpt.placeholder = 'deepseek-chat';
		fallbackOpt.default = 'deepseek-chat';

		o = s.option(form.Value, 'max_iterations', _('Pasos / Iteraciones Máximas (Anti-Loop)'),
			_('Límite máximo de herramientas consecutivas ejecutadas por consulta (default: 5).'));
		o.datatype = 'uinteger';
		o.default = '5';

		/* ── 6. Visor de Logs en Tiempo Real (Live Logs) ── */
		s = m.section(form.NamedSection, '_logs', 'clawrt', _('Visor de Registros del Agente (Live Logs)'));
		s.anonymous = true;
		s.addremove = false;

		o = s.option(form.DummyValue, '_logs_view', _('Últimas Entradas del Sistema (logread -e clawrt)'));
		o.cfgvalue = function() {
			return '<pre style="background:#1e1e1e; color:#d4d4d4; padding:12px; border-radius:6px; max-height:220px; overflow-y:auto; font-family:monospace; font-size:12px;">' +
				logsOutput.replace(/</g, '&lt;').replace(/>/g, '&gt;') +
				'</pre>';
		};
		o.rawhtml = true;

		return m.render().then(function(node) {
			setTimeout(function() {
				var provSelect = node.querySelector('select[id$="cbid.clawrt.llm.provider"]') || node.querySelector('select[name$="llm.provider"]');
				var urlInput = node.querySelector('input[id$="cbid.clawrt.llm.base_url"]') || node.querySelector('input[name$="llm.base_url"]');
				var modelInput = node.querySelector('input[id$="cbid.clawrt.llm.model"]') || node.querySelector('input[name$="llm.model"]');
				var fallbackInput = node.querySelector('input[id$="cbid.clawrt.llm.fallback_model"]') || node.querySelector('input[name$="llm.fallback_model"]');

				if (provSelect) {
					provSelect.addEventListener('change', function() {
						var val = provSelect.value;
						var defaults = providerDefaults[val] || providerDefaults['custom'];
						if (urlInput && defaults.url) {
							urlInput.value = defaults.url;
							urlInput.placeholder = defaults.url;
						}
						if (modelInput && defaults.model) {
							modelInput.placeholder = defaults.model;
							if (!modelInput.value || modelInput.value === 'gpt-4o-mini') {
								modelInput.value = defaults.model;
							}
						}
						if (fallbackInput && defaults.fallback) {
							fallbackInput.placeholder = defaults.fallback;
							if (!fallbackInput.value || fallbackInput.value === 'deepseek-chat') {
								fallbackInput.value = defaults.fallback;
							}
						}
					});
				}

				// Interactive Buttons Event Handlers
				var btnLLM = document.getElementById('btn_test_llm');
				if (btnLLM) {
					btnLLM.addEventListener('click', function() {
						ui.showModal(_('Probando Conexión LLM...'), [
							E('p', { class: 'spinning' }, _('Enviando consulta de prueba de 1 token al proveedor de IA...'))
						]);
						fs.exec('/usr/bin/clawrt', ['-test-llm']).then(function(res) {
							var out = res && res.stdout ? res.stdout : (res && res.stderr ? res.stderr : 'Sin respuesta');
							var isOk = res && res.code === 0;
							ui.showModal(isOk ? _('✅ Prueba LLM Exitosa') : _('❌ Fallo en Prueba LLM'), [
								E('p', {}, out),
								E('div', { class: 'right' }, [
									E('button', { class: 'btn', click: ui.hideModal }, _('Cerrar'))
								])
							]);
						});
					});
				}

				var btnFetch = document.getElementById('btn_fetch_models');
				if (btnFetch) {
					btnFetch.addEventListener('click', function() {
						ui.showModal(_('Detectando Modelos en Vivo...'), [
							E('p', { class: 'spinning' }, _('Consultando el endpoint /v1/models del proveedor configurado...'))
						]);
						fs.exec('/usr/bin/clawrt', ['-fetch-models']).then(function(res) {
							var out = res && res.stdout ? res.stdout : (res && res.stderr ? res.stderr : 'Sin respuesta');
							var isOk = res && res.code === 0;
							ui.showModal(isOk ? _('🔍 Modelos Disponibles en Vivo') : _('⚠️ Error al consultar /v1/models'), [
								E('p', {}, isOk ? _('Modelos activos devueltos por el proveedor:') : _('No se pudieron obtener los modelos. Revisa tu API Key o conectividad.')),
								E('pre', { style: 'background:#1e1e1e; color:#d4d4d4; padding:10px; border-radius:4px; max-height:200px; overflow-y:auto;' }, out),
								E('div', { class: 'right' }, [
									E('button', { class: 'btn', click: ui.hideModal }, _('Cerrar'))
								])
							]);
						});
					});
				}

				var btnTg = document.getElementById('btn_test_telegram');
				if (btnTg) {
					btnTg.addEventListener('click', function() {
						ui.showModal(_('Probando Bot de Telegram...'), [
							E('p', { class: 'spinning' }, _('Verificando token y enviando notificación de prueba a Telegram...'))
						]);
						fs.exec('/usr/bin/clawrt', ['-test-telegram']).then(function(res) {
							var out = res && res.stdout ? res.stdout : (res && res.stderr ? res.stderr : 'Sin respuesta');
							var isOk = res && res.code === 0;
							ui.showModal(isOk ? _('✅ Prueba Telegram Exitosa') : _('❌ Fallo en Telegram'), [
								E('p', {}, out),
								E('div', { class: 'right' }, [
									E('button', { class: 'btn', click: ui.hideModal }, _('Cerrar'))
								])
							]);
						});
					});
				}

				var btnClear = document.getElementById('btn_clear_facts');
				if (btnClear) {
					btnClear.addEventListener('click', function() {
						fs.exec('/usr/bin/clawrt', ['-clear-facts']).then(function(res) {
							ui.showModal(_('🧹 Memoria Limpiada'), [
								E('p', {}, _('Los hechos dinámicos y caché aprendidos en /tmp/clawrt_facts.json han sido eliminados.')),
								E('div', { class: 'right' }, [
									E('button', { class: 'btn', click: ui.hideModal }, _('Cerrar'))
								])
							]);
						});
					});
				}
			}, 200);

			return node;
		});
	},

	handleSave: function(ev) {
		return this.super('handleSave', [ev]).then(function() {
			return fs.exec('/etc/init.d/clawrt', ['reload']).catch(function() {});
		});
	},
});
