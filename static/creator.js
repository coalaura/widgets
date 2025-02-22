(async () => {
	const Widgets = await fetch("/widgets.json").then(response => response.json()),
		KnownOptions = {
			color: def => input("color", "color", def),
			align: def => select("align", def, ["left", "center", "right"]),
			weight: def => select("weight", def, ["100", "200", "300", "400", "500", "600", "700", "800", "900"]),
		};

	const $page = document.getElementById("page");

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
	function build(name, options = {}) {
		const query = new URLSearchParams(options).toString();

		return `${window.location.origin}/${name}${query ? `?${query}` : ""}`;
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

	function footer() {
		const end = new Date().getFullYear().toString();

		return `<div id="footer">
			<div>&copy; 2025${end !== "2025" ? ` - ${end}` : ""} <a href="https://github.com/coalaura" target="_blank">coalaura</a></div>
			<a href="/" id="home" title="Back to widget list.">🏠</a>
		</div>`;
	}

	function opt(name, def) {
		let field;

		if (name in KnownOptions) {
			field = KnownOptions[name](def);
		} else {
			field = input("text", name, def);
		}

		return `<div class="option">
			<label for="opt_${name}">${ucfirst(name)}</label>
			${field}
			<button class="undo square" title="Reset value to default.">✖</button>
		</div>`;
	}

	function show(widget) {
		window.location.hash = widget.name;

		const options = {};

		let html = `<h1>
			${ucfirst(widget.name)}

			<button id="export" class="square" title="Copy widget url with customizations.">❐</button>
		</h1><div id="options">`;

		for (const name in widget.options) {
			const def = widget.options[name];

			options[name] = def;

			html += opt(name, def);
		}

		html += `</div><iframe id="preview" src=""></iframe>`;

		$page.innerHTML = html + footer();

		const $export = document.getElementById("export"),
			$options = [...document.querySelectorAll(".option")],
			$preview = document.getElementById("preview");

		let timeout;

		$export.addEventListener("click", async function() {
			clearTimeout(timeout);

			try {
				const url = build(widget.name, options);

				await navigator.clipboard.writeText(url);
			} catch {
				return;
			}

			this.classList.add("copied");

			timeout = setTimeout(() => {
				this.classList.remove("copied");
			}, 750);
		}, false);

		for (const $opt of $options) {
			const $undo = $opt.querySelector(".undo"),
				$input = $opt.querySelector("input,select");

			const name = $input.name;

			$input.addEventListener("input", function() {
				set($undo, name, this.value);
			}, false);

			$undo.addEventListener("click", function() {
				const value = widget.options[name];

				$input.value = value;

				set(this, name, value);
			}, false);
		}

		function update() {
			const clean = {};

			for (const name in options) {
				const value = options[name];

				if (value && widget.options[name] !== value) {
					clean[name] = value;
				}
			}

			const url = build(widget.name, clean);

			if ($preview.src !== url) {
				$preview.src = url;
			}
		}

		function set($undo, name, value) {
			if (widget.options[name] === value) {
				$undo.classList.remove("changed");
			} else {
				$undo.classList.add("changed");
			}

			options[name] = value;

			update();
		}

		update();
	}

	function render() {
		let html = `<h1>Widgets</h1><div id="widgets">`;

		for (const widget of Widgets) {
			const url = build(widget.name, {
				color: "#cad3f5",
			});

			html += `<div class="widget" data-name="${widget.name}">
				<div class="header">
					<div class="name">${ucfirst(widget.name)}</div>
					<div class="description">${widget.description}</div>
				</div>
				<iframe src="${url}"></iframe>
			</div>`;
		}

		html += "</div>";

		$page.innerHTML = html + footer();

		document.querySelectorAll(".widget").forEach(el => {
			const name = el.dataset.name,
				widget = Widgets.find(w => w.name === name);

			el.addEventListener(
				"click",
				() => {
					show(widget);
				},
				false
			);
		});
	}
})();
