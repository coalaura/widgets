(async () => {
	const $page = document.getElementById("page"),
		appearance = ["font", "size", "weight", "color", "align", "fps"],
		calendarTokens = [
			["YYYY", "2026", "Four-digit year"], ["YY", "26", "Two-digit year"],
			["MMMM", "October", "Full month"], ["MMM", "Oct", "Short month"],
			["MM", "10", "Padded month"], ["M", "10", "Month number"],
			["DD", "03", "Padded day"], ["D", "3", "Day of month"], ["Do", "3rd", "Ordinal day"],
			["dddd", "Saturday", "Full weekday"], ["ddd", "Sat", "Short weekday"],
			["dd", "Sa", "Two-letter weekday"], ["d", "6", "Weekday number (Sunday = 0)"]
		],
		timeTokens = [
			["HH", "14", "24-hour, padded"], ["H", "14", "24-hour"],
			["hh", "02", "12-hour, padded"], ["h", "2", "12-hour"],
			["mm", "05", "Padded minute"], ["m", "5", "Minute"],
			["ss", "09", "Padded second"], ["s", "9", "Second"],
			["SSS", "123", "Milliseconds"], ["A", "PM", "Uppercase AM/PM"], ["a", "pm", "Lowercase am/pm"],
			["Z", "+02:00", "UTC offset"], ["ZZ", "+0200", "Compact UTC offset"]
		],
		formatGuides = {
			date: {
				format: {
					title: "Date & time format",
					previewLabel: "Current local date and time",
					examples: ["dddd, MMMM D, h:mm A", "YYYY-MM-DD", "ddd, MMM D", "HH:mm"],
					tokens: [...calendarTokens.filter(([token]) => token !== "Do"), ...timeTokens]
				}
			},
			holiday: {
				date_format: {
					title: "Holiday date format",
					previewLabel: "Example: October 3, 2026",
					utc: true,
					examples: ["MMM Do", "DD.MM.YYYY", "dddd, MMMM D", "YYYY-MM-DD"],
					tokens: calendarTokens.filter(([token]) => token !== "d" && token !== "dd")
				}
			}
		};

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
		return name.charAt(0).toUpperCase() + name.slice(1).replaceAll("_", " ");
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

	function opt(name, option, guide) {
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
				<div class="option-actions">
					${guide ? `<button class="format-help" type="button" data-option="${name}" aria-label="Help with ${ucfirst(name)}" title="Format guide">?</button>` : ""}
					<button class="undo" type="button" title="Reset ${name} to default" aria-label="Reset ${name} to default">Reset</button>
				</div>
			</div>
			${field}
			<div class="description">${description}</div>
		</div>`;
	}

	function previewDateFormat(pattern, guide) {
		const date = guide.utc ? new Date("2026-10-03T12:00:00Z") : new Date(),
			part = field => date[`get${guide.utc ? "UTC" : ""}${field}`](),
			pad = value => String(value).padStart(2, "0"),
			year = part("FullYear"),
			month = part("Month") + 1,
			day = part("Date"),
			weekday = part("Day"),
			hour = part("Hours"),
			minute = part("Minutes"),
			second = part("Seconds"),
			intlOptions = guide.utc ? { timeZone: "UTC" } : {},
			monthName = new Intl.DateTimeFormat("en", { ...intlOptions, month: "long" }).format(date),
			weekdayName = new Intl.DateTimeFormat("en", { ...intlOptions, weekday: "long" }).format(date),
			lastDigit = day % 10,
			suffix = day % 100 >= 11 && day % 100 <= 13 ? "th" :
				lastDigit === 1 ? "st" : lastDigit === 2 ? "nd" : lastDigit === 3 ? "rd" : "th",
			offset = guide.utc ? 0 : -date.getTimezoneOffset(),
			zone = `${offset < 0 ? "-" : "+"}${pad(Math.floor(Math.abs(offset) / 60))}:${pad(Math.abs(offset) % 60)}`,
			parts = {
				YYYY: year, YY: String(year).slice(-2),
				MMMM: monthName, MMM: new Intl.DateTimeFormat("en", { ...intlOptions, month: "short" }).format(date),
				MM: pad(month), M: month,
				DD: pad(day), D: day, Do: `${day}${suffix}`,
				dddd: weekdayName, ddd: new Intl.DateTimeFormat("en", { ...intlOptions, weekday: "short" }).format(date),
				dd: weekdayName.slice(0, 2), d: weekday,
				HH: pad(hour), H: hour, hh: pad(hour % 12 || 12), h: hour % 12 || 12,
				mm: pad(minute), m: minute, ss: pad(second), s: second, SSS: String(date.getMilliseconds()).padStart(3, "0"),
				A: hour < 12 ? "AM" : "PM", a: hour < 12 ? "am" : "pm", Z: zone, ZZ: zone.replace(":", "")
			},
			tokens = guide.tokens.map(([token]) => token).sort((first, second) => second.length - first.length),
			matcher = new RegExp(`\\[[^\\]]*\\]|${tokens.join("|")}`, "g");

		return pattern.replace(matcher, token => token.startsWith("[") ? token.slice(1, -1) : parts[token]);
	}

	function formatDialog() {
		return `<dialog id="format-dialog" aria-labelledby="format-title" aria-describedby="format-intro">
			<div class="format-dialog-heading">
				<h2 id="format-title"></h2>
				<button class="format-close" type="button" aria-label="Close format guide">×</button>
			</div>
			<p id="format-intro">Choose a pattern or click tokens to insert them at the cursor. Wrap literal words in [brackets].</p>
			<label for="format-builder">Your format</label>
			<input id="format-builder" type="text" spellcheck="false" />
			<div class="format-preview">
				<div id="format-preview-label"></div>
				<output id="format-preview" for="format-builder" aria-live="polite"></output>
			</div>
			<div class="format-dialog-scroll">
				<h3>Examples</h3>
				<div id="format-examples" class="format-examples"></div>
				<h3>Available tokens</h3>
				<div id="format-tokens" class="format-tokens"></div>
			</div>
			<div class="format-dialog-actions">
				<button class="format-cancel" type="button">Cancel</button>
				<button class="format-apply" type="button">Use format</button>
			</div>
		</dialog>`;
	}

	function connectFormatDialog(widget) {
		const $dialog = document.getElementById("format-dialog"),
			$builder = $dialog.querySelector("#format-builder"),
			$preview = $dialog.querySelector("#format-preview");
		let $input, guide;

		function updatePreview() {
			$preview.textContent = $builder.value ? previewDateFormat($builder.value, guide) : "Enter a format to see a preview";
		}

		$builder.addEventListener("input", updatePreview);

		for (const $button of document.querySelectorAll(".format-help")) {
			$button.addEventListener("click", () => {
				$input = document.getElementById(`opt_${$button.dataset.option}`);
				guide = formatGuides[widget.name][$button.dataset.option];
				$builder.value = $input.value;
				$dialog.querySelector("#format-title").textContent = guide.title;
				$dialog.querySelector("#format-preview-label").textContent = guide.previewLabel;
				$dialog.querySelector("#format-examples").innerHTML = guide.examples.map((pattern, index) =>
					`<button type="button" data-example="${index}"><code>${pattern}</code></button>`).join("");
				$dialog.querySelector("#format-tokens").innerHTML = guide.tokens.map(([token, example, description], index) =>
					`<button type="button" data-token="${index}" title="Insert ${token}"><code>${token}</code><span>${description}</span><em>${example}</em></button>`).join("");
				updatePreview();
				$dialog.showModal();
				$builder.focus();
			});
		}

		$dialog.addEventListener("click", event => {
			const $example = event.target.closest("[data-example]"),
				$token = event.target.closest("[data-token]");

			if ($example) {
				$builder.value = guide.examples[Number($example.dataset.example)];
				$builder.focus();
				$builder.setSelectionRange($builder.value.length, $builder.value.length);
				updatePreview();
			} else if ($token) {
				const token = guide.tokens[Number($token.dataset.token)][0];

				$builder.setRangeText(token, $builder.selectionStart, $builder.selectionEnd, "end");
				$builder.focus();
				updatePreview();
			}
		});

		$dialog.querySelector(".format-apply").addEventListener("click", () => {
			$input.value = $builder.value;
			$input.dispatchEvent(new Event("input", { bubbles: true }));
			$dialog.close();
			$input.focus();
		});

		for (const selector of [".format-close", ".format-cancel"]) {
			$dialog.querySelector(selector).addEventListener("click", () => $dialog.close());
		}
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
				specificOptions += opt(name, widget.options[name], formatGuides[widget.name]?.[name]);
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

		$page.innerHTML = html + (formatGuides[widget.name] ? formatDialog() : "") + footer();

		if (formatGuides[widget.name]) {
			connectFormatDialog(widget);
		}

		const $export = document.getElementById("export"),
			$import = document.getElementById("import-form"),
			$options = [...document.querySelectorAll(".option")],
			$preview = document.getElementById("preview"),
			$stage = document.querySelector(".preview-stage"),
			$holidayDisplay = widget.name === "holiday" ? document.getElementById("opt_display") : null;

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

		if ($holidayDisplay) {
			const $dateFormat = document.getElementById("opt_date_format").closest(".option");
			$holidayDisplay.closest(".option").after($dateFormat);

			function showDateFormat() {
				$dateFormat.hidden = $holidayDisplay.value === "relative";
			}

			$holidayDisplay.addEventListener("input", showDateFormat);
			showDateFormat();
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
