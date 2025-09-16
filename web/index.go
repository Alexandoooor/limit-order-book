package web

import (
	"html/template"
	"net/http"
	"os"
	"strings"
)

var indexTmpl = template.Must(template.New("index").Parse(IndexHTML))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "unknown"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]string{"Hostname": hostname}

	if err := indexTmpl.Execute(w, data); err != nil {
		// fallback: literal replace to ensure hostname is visible even if template engine chokes
		fallback := strings.ReplaceAll(IndexHTML, "{{.Hostname}}", hostname)
		_, _ = w.Write([]byte(fallback))
	}
}

func IndexTemplate() string {
	return IndexHTML
}

// IndexHTML contains the minimal UI served at GET /
const IndexHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>Limit Order Book</title>
	<style>
	:root { font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial; }
	body { margin: 0; padding: 24px; background: #071024; color: #e6eef6; }
	.card { max-width: 1000px; margin: 0 auto; background: #091426; padding: 16px; border-radius: 12px; }
	.row { display:flex; gap:8px; align-items:center; }
	input, select, button { padding:8px; border-radius:8px; background:#071426; border:1px solid #1f2b3a; color:#c9d8e6 }
	table { width:100%; margin-top:12px; border-collapse:collapse }
	th, td { padding:8px; border-bottom:1px solid #123047; text-align:left }
	.up{ color:#22c55e } .down{ color:#f43f5e }
	</style>
</head>
<body>
	<div class="card">
	<h2>Limit Order Book</h2>

	<div class="row">
	  <strong>New Connected Pod:</strong> <span>{{.Hostname}}</span>
	</div>


	<h2>Place Order</h2>
	<div class="row">
	<form id="orderForm">
		<label>Side:</label>
		<select id="side" name="side">
			<option value="0">Buy</option>
			<option value="1">Sell</option>
		</select><br>
		<label>Price:</label>
		<input type="number" id="price" name="price" required><br>
		<label>Size:</label>
		<input type="number" id="size" name="size" required><br>
		<button type="submit">Place Order</button>
	</form>
	</div>

	<div class="row">
		<div class="column">
			<h2>Bids</h2>
			<table>
				<tr><th>Price</th><th>Volume</th></tr>
				{{range .Bids}}
				<tr>
					<td>{{.Price}}</td>
					<td>{{.Volume}}</td>
				</tr>
				{{end}}
			</table>
		</div>

		<div class="column">
		<h2>Asks</h2>
		<table>
			<tr><th>Price</th><th>Volume</th></tr>
			{{range .Asks}}
			<tr>
				<td>{{.Price}}</td>
				<td>{{.Volume}}</td>
			</tr>
			{{end}}
		</table>
		</div>
	</div>


	<script>
		const form = document.getElementById('orderForm');
		form.addEventListener('submit', async (e) => {
			e.preventDefault();
			const side = parseInt(document.getElementById('side').value);
			const price = parseInt(document.getElementById('price').value);
			const size = parseInt(document.getElementById('size').value);

			const resp = await fetch('/api/order', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify({side, price, size})
			});
			if (resp.ok) {
				window.location.reload(); // reload to refresh order book
			} else {
				const text = await resp.text();
				alert('Error: ' + text);
			}
		});
	</script>
	</div>
</body>
</html>`
