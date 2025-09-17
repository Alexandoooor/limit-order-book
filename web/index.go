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
	:root{
	--bg:#0f1724;
	--card:#0b1220;
	--muted:#9aa4b2;
	--accent:linear-gradient(90deg,#06b6d4,#7c3aed);
	--glass: rgba(255,255,255,0.03);
	--success:#16a34a;
	--danger:#ef4444;
	--radius:16px;
	font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
	}
	*{box-sizing:border-box}
	html,body{height:100%}
	body{
	margin:0; background: radial-gradient(1200px 600px at 10% 10%, rgba(124,58,237,0.12), transparent),
	radial-gradient(900px 400px at 90% 90%, rgba(6,182,212,0.06), transparent), var(--bg);
	color:#e6eef6; display:flex; align-items:center; justify-content:center; padding:24px;
	}

	.card{
	width:100%;
	max-width:1100px;
	background:linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01));
	border-radius:var(--radius);
	padding:32px;
	box-shadow: 0 6px 30px rgba(2,6,23,0.7);
	border:1px solid rgba(255,255,255,0.03);
	}

	h1{font-size:22px;margin:0 0 8px 0}
	p.lead{margin:0 0 24px 0;color:var(--muted);font-size:14px}

	form{
		display:grid;
		gap:5px;
		font-size:15px;
	}

	.row{display:flex; gap:12px}

	label{
		font-size:15px;
		color:var(--muted);
		display:block;
		margin-bottom:6px
	}

	select,input[type="number"]{
	width:100%; padding:12px 14px; background:var(--glass); border:1px solid rgba(255,255,255,0.04); color:inherit;
	border-radius:10px; font-size:15px; outline:none;
	}
	select:focus,input[type="number"]:focus{box-shadow:0 4px 18px rgba(124,58,237,0.08); border-color:rgba(124,58,237,0.28)}

	.side-by-side{display:grid; grid-template-columns: 1fr 1fr; gap:20px}

	.order-type{
	flex:1; display:flex; align-items:center; gap:12px; padding:12px 16px; border-radius:12px; cursor:pointer;
	user-select:none; border:1px solid rgba(255,255,255,0.03); background:transparent; font-size:15px;
	}
	.order-type input{display:none}

	.order-type.buy{background:linear-gradient(90deg, rgba(22,163,74,0.08), transparent); border-color: rgba(16,185,129,0.08)}
	.order-type.sell{background:linear-gradient(90deg, rgba(239,68,68,0.06), transparent); border-color: rgba(239,68,68,0.08)}

	.price-row{display:flex; align-items:center; gap:8px}
	.currency{padding:8px 10px; border-radius:8px; background:rgba(255,255,255,0.02); font-size:14px; color:var(--muted); border:1px solid rgba(255,255,255,0.02)}

	.meta{display:flex; justify-content:space-between; align-items:center; font-size:14px; color:var(--muted)}

	.total{font-weight:600; font-size:16px}

	button.submit{
	width:100%; max-width:400px; margin:0 auto; display:block;
	padding:14px 16px; border-radius:12px; border:1px solid rgba(255,255,255,0.04); font-weight:600; font-size:15px; cursor:pointer;
	background:var(--glass); color:inherit; text-align:center;
	box-shadow: inset 0 0 0 0 transparent;
	}
	button.submit.buy{background:linear-gradient(90deg, rgba(22,163,74,0.08), transparent); border-color: rgba(16,185,129,0.08); color:#10b981}
	button.submit.sell{background:linear-gradient(90deg, rgba(239,68,68,0.06), transparent); border-color: rgba(239,68,68,0.08); color:#ef4444}
	button:disabled{opacity:0.45; cursor:not-allowed}

	.hint{font-size:13px;color:var(--muted); text-align:center}

	@media (max-width:900px){
	.side-by-side{grid-template-columns:1fr}
	.card{text-align:left; align-items:stretch}
	form{max-width:100%}
	button.submit{max-width:100%}
	}
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
			<option value="buy">Buy</option>
			<option value="sell">Sell</option>
		</select><br>
		<label>Price:</label>
		<input type="number" id="price" name="price" required><br>
		<label>Size:</label>
		<input type="number" id="size" name="size" required><br>
        	<button type="submit" id="submitBtn" class="submit buy order-type buy">Place Order</button>
	</form>
	</div>

	<div class="column">
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
	</div>

	<div class="column">
		<h2>Trades</h2>
		<table>
			<tr><th>Price</th><th>Size</th><th>Time</th><th>Buyer ID</th><th>Seller ID</th></tr>
			{{range .Trades}}
			<tr>
				<td>{{.Price}}</td>
				<td>{{.Size}}</td>
				<td>{{.Time}}</td>
				<td>{{.BuyerID}}</td>
				<td>{{.SellerID}}</td>
			</tr>
			{{end}}
		</table>

	</div>

	<script>
		const form = document.getElementById('orderForm');
		form.addEventListener('submit', async (e) => {
			e.preventDefault();
			const side = document.getElementById('side').value;
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
