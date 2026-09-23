(async () => {
	const $page = document.getElementById("page"),
		appearance = ["font", "size", "weight", "color", "align", "fps"];

	let Widgets, listObserver;

	try {
		const response = await fetch("/widgets.json");

		if (!response.ok) {
			throw new Error(`Unable to load widgets (${response.status})`);
		}

		Widgets = await response.json();
	} catch {
		$page.innerHTML = `<div class="page-heading"><div><div class="eyebrow">Your workspace</div><h1>Widgets</h1></div></div><div class="catalog-loading" role="alert">Couldn't load widgets. Please refresh to try again.</div>`;

		return;
	}

	// Return to previous page
	{
		const name = window.location.hash.substring(1),
			widget = Widgets.find(w => w.name === name);

		if (widget) {
			show(widget);
		} else {
			render();
		}
	}

	// Functions
	function build(widget, options = {}) {
		const cleaned = {};

		for (const [key, value] of Object.entries(options)) {
			const option = widget.options[key],
				def = option.default;

			if (value !== def) {
				cleaned[key] = value;
			}
		}

		const query = new URLSearchParams(cleaned).toString();

		return `${window.location.origin}/${widget.name}${query ? `?${query}` : ""}`;
	}

	function ucfirst(name) {
		return name.charAt(0).toUpperCase() + name.slice(1);
	}

	function input(type, name, value) {
		return `<input type="${type}" value="${value}" name="${name}" id="opt_${name}" />`;
	}

	function select(name, selected, options) {
		const opts = options.map(o => `<option value="${o}"${selected === o ? "selected" : ""}>${o}</option>`).join("");

		return `<select name="${name}" id="opt_${name}">${opts}</select>`;
	}

	function toggle(name, active) {
		return `<select name="${name}" id="opt_${name}">
			<option value="1" ${active ? "selected" : ""}>On</option>
			<option value="0" ${!active ? "selected" : ""}>Off</option>
		</select>`;
	}

	function footer(showHome = true) {
		const end = new Date().getFullYear().toString();

		return `<div id="footer">
			<div>&copy; 2025${end !== "2025" ? ` - ${end}` : ""} <a href="https://github.com/coalaura" target="_blank">coalaura</a></div>
			${showHome ? `<a href="/" id="home" title="Back to widget list">All widgets ↑</a>` : ""}
		</div>`;
	}

	function opt(name, option) {
		const { type, default: def, description, allowed } = option;

		let field;

		if (type === "toggle") {
			field = toggle(name, def);
		} else if (type === "select") {
			field = select(name, def, allowed);
		} else {
			field = input(type, name, def);
		}

		return `<div class="option">
			<div class="option-heading">
				<label for="opt_${name}">${ucfirst(name)}</label>
				<button class="undo" type="button" title="Reset ${name} to default" aria-label="Reset ${name} to default">Reset</button>
			</div>
			${field}
			<div class="description">${description}</div>
		</div>`;
	}

	function importUrl(value) {
		const url = new URL(value.trim(), window.location.origin);

		if (url.protocol !== "http:" && url.protocol !== "https:") {
			throw new Error("Enter an HTTP or HTTPS widget URL.");
		}

		const name = url.pathname.replace(/^\/|\/$/g, ""),
			widget = Widgets.find(item => item.name === name);

		if (!widget) {
			throw new Error("No widget matches the path in that URL.");
		}

		const settings = {};

		for (const [name, option] of Object.entries(widget.options)) {
			const raw = url.searchParams.get(name);

			if (raw === null || raw === "") {
				continue;
			}

			if (option.type === "select" && !option.allowed.includes(raw)) {
				throw new Error(`Invalid ${name} value in URL.`);
			}

			if (option.type === "number") {
				const field = document.createElement("input");

				field.type = "number";
				field.value = raw;

				if (field.value === "" || !Number.isFinite(Number(raw))) {
					throw new Error(`Invalid ${name} value in URL.`);
				}
			}

			if (option.type === "color" && !/^#([\da-f]{3}|[\da-f]{6})$/i.test(raw)) {
				throw new Error(`${ucfirst(name)} must be a hex color to edit here.`);
			}

			if (option.type === "toggle") {
				settings[name] = ["1", "ok", "true"].includes(raw.toLowerCase()) ? "1" : "0";
			} else if (option.type === "color" && raw.length === 4) {
				settings[name] = `#${raw[1]}${raw[1]}${raw[2]}${raw[2]}${raw[3]}${raw[3]}`;
			} else {
				settings[name] = raw;
			}
		}

		return { widget, settings };
	}

	function show(widget, settings = {}, importedUrl = "") {
		listObserver?.disconnect();
		window.location.hash = widget.name;

		const options = {};

		let specificOptions = "",
			appearanceOptions = "";

		for (const name in widget.options) {
			options[name] = widget.options[name].default;

			if (!appearance.includes(name)) {
				specificOptions += opt(name, widget.options[name]);
			}
		}

		for (const name of appearance) {
			if (widget.options[name]) {
				appearanceOptions += opt(name, widget.options[name]);
			}
		}

		const html = `<div class="page-heading">
			<div><div class="eyebrow">Customize widget</div><h1>${ucfirst(widget.name)}</h1></div>
			<button id="export" type="button" title="Copy widget URL" aria-live="polite">Copy URL</button>
		</div>
		<div id="editor">
			<div id="options">
				<details class="import-panel">
					<summary>Load from URL</summary>
					<form id="import-form">
						<label for="import-url">Widget URL</label>
						<div class="import-controls">
							<input id="import-url" type="text" placeholder="https://example.com/widget?color=%23..." autocomplete="off" spellcheck="false" required />
							<button type="submit">Load settings</button>
						</div>
						<p id="import-message" role="status">Paste a widget URL; its host will be ignored.</p>
					</form>
				</details>
				${specificOptions ? `<section class="option-group"><h2>Widget settings</h2><div class="option-grid">${specificOptions}</div></section>` : ""}
				<section class="option-group"><h2>Appearance</h2><div class="option-grid">${appearanceOptions}</div></section>
			</div>
			<section class="preview-panel">
				<div class="preview-heading"><h2>Live preview</h2><span>Updates as you edit</span></div>
				<div class="preview-stage is-loading">
					<span class="preview-status" role="status">Loading preview…</span>
					<iframe id="preview" class="${widget.size || ""}" title="${widget.name} preview"></iframe>
				</div>
			</section>
		</div>`;

		$page.innerHTML = html + footer();

		const $export = document.getElementById("export"),
			$import = document.getElementById("import-form"),
			$options = [...document.querySelectorAll(".option")],
			$preview = document.getElementById("preview"),
			$stage = document.querySelector(".preview-stage");

		$import.querySelector("input").value = importedUrl;

		if (importedUrl) {
			$import.closest("details").open = true;

			$import.querySelector("#import-message").textContent = "Settings loaded from URL.";
		}

		$import.addEventListener("submit", event => {
			event.preventDefault();

			const value = $import.querySelector("input").value;

			try {
				const loaded = importUrl(value);

				show(loaded.widget, loaded.settings, value);
			} catch (error) {
				const $message = $import.querySelector("#import-message");

				$message.textContent = error.message;
				$message.classList.add("error");
			}
		});

		$preview.addEventListener("load", () => {
			if ($preview.hasAttribute("src")) {
				$stage.classList.remove("is-loading");
			}
		});

		let timeout;

		$export.addEventListener(
			"click",
			async function () {
				clearTimeout(timeout);

				try {
					const url = build(widget, options);

					await navigator.clipboard.writeText(url);
				} catch {
					return;
				}

				this.textContent = "Copied!";
				this.classList.add("copied");

				timeout = setTimeout(() => {
					this.textContent = "Copy URL";
					this.classList.remove("copied");
				}, 1600);
			},
			false
		);

		for (const $opt of $options) {
			const $undo = $opt.querySelector(".undo"),
				$input = $opt.querySelector("input,select");

			const name = $input.name;

			if (Object.hasOwn(settings, name)) {
				$input.value = settings[name];

				set($undo, name, settings[name], false);
			}

			$input.addEventListener(
				"input",
				function () {
					set($undo, name, this.value);
				},
				false
			);

			$undo.addEventListener(
				"click",
				function () {
					const option = widget.options[name],
						def = option.default;

					$input.value = def;

					set(this, name, def);
				},
				false
			);
		}

		function update() {
			const url = build(widget, options);

			if ($preview.src !== url) {
				$stage.classList.add("is-loading");

				$preview.src = url;
			}
		}

		function normalize(option, value) {
			switch (option.type) {
				case "number":
					return Number(value);
				case "toggle":
					return value === true || String(value) === "1";
			}

			return String(value);
		}

		function set($undo, name, value, refresh = true) {
			const option = widget.options[name];

			if (normalize(option, option.default) === normalize(option, value)) {
				$undo.classList.remove("changed");
			} else {
				$undo.classList.add("changed");
			}

			options[name] = value;

			if (refresh) {
				update();
			}
		}

		update();
	}

	function render() {
		listObserver?.disconnect();

		let html = `<div class="page-heading">
			<div><div class="eyebrow">Your workspace</div><h1>Widgets</h1><p>Pick a widget to make it yours.</p></div>
			<span class="widget-count">${Widgets.length} available</span>
		</div><div id="widgets">`;

		for (const widget of Widgets) {
			const url = build(widget);

			html += `<div class="widget" data-name="${widget.name}" role="button" tabindex="0" aria-label="Customize ${widget.name}">
				<div class="widget-preview preview-stage is-loading ${widget.size || ""}">
					<span class="preview-status" role="status">Loading preview…</span>
					<iframe data-src="${url}" loading="lazy" title="${widget.name} preview" class="${widget.size || ""}" tabindex="-1"></iframe>
				</div>
				<div class="widget-info">
					<div class="widget-title"><span>${ucfirst(widget.name)}</span><span aria-hidden="true">↗</span></div>
					<p class="description">${widget.description}</p>
				</div>
			</div>`;
		}

		html += "</div>";

		$page.innerHTML = html + footer(false);

		const $list = document.getElementById("widgets"),
			$cards = [...$list.querySelectorAll(".widget")];

		listObserver = "IntersectionObserver" in window
			? new IntersectionObserver(entries => {
				for (const entry of entries) {
					if (entry.isIntersecting) {
						const frame = entry.target.querySelector("iframe");

						frame.src = frame.dataset.src;

						listObserver.unobserve(entry.target);
					}
				}
			}, { root: $list, rootMargin: "200px" })
			: null;

		for (const el of $cards) {
			const name = el.dataset.name,
				widget = Widgets.find(w => w.name === name),
				frame = el.querySelector("iframe");

			el.addEventListener("click", () => show(widget), false);

			el.addEventListener("keydown", event => {
				if (event.key === "Enter" || event.key === " ") {
					event.preventDefault();
					show(widget);
				}
			});

			frame.addEventListener("load", () => {
				if (frame.hasAttribute("src")) {
					el.querySelector(".preview-stage").classList.remove("is-loading");
				}
			});

			if (listObserver) {
				listObserver.observe(el);
			} else {
				frame.src = frame.dataset.src;
			}
		}
	}

	window.addEventListener("popstate", () => {
		if (window.location.hash) {
			const widget = Widgets.find(w => w.name === window.location.hash.substring(1));

			if (widget) {
				show(widget);
			} else {
				window.location.hash = "";
			}

			return;
		}

		render();
	});
})();
