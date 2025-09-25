package web

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

	.container { display: flex; gap: 2rem; }
	.order-form { flex: 0 0 350px; gap: 2rem; margin-bottom: 6px; }
	.order-book { flex: 1; display: flex; flex-direction: column; gap: 2rem; margin-bottom: 6px; }

	.card{
	width:100%;
	max-width:950px;
	background:linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01));
	border-radius:var(--radius);
	padding:32px;
	box-shadow: 0 6px 30px rgba(2,6,23,0.7);
	border:1px solid rgba(255,255,255,0.03);
	}

	h1{font-size:1.5rem; margin:0 0 8px 0; text-align: center;}
	h2{margin: 0 0 8px 0; font-size: 1.125rem; color: #fff; }
	p.lead{margin:0 0 24px 0;color:var(--muted);font-size:14px}

	form{
		display:grid;
		gap:20px;
		font-size:15px;
	}

	.row{display:flex; gap:12px}
	table {
		width: 100%;
		border-collapse: collapse;
		background-color: #1e1e2f;
		color: #fff;
		border-radius:12px;
		overflow: hidden;
		border: 1px solid;
	}
	th, td {
		padding: 8px 12px;
		text-align: left;
		border-bottom: 1px solid #333;
		border-right: 1px solid #333;
		font-size: 15px;
		font-weight: 400;
		font-family: inherit;
	}
	th {
		background-color: #2c2c3f;
	}
	td {
		color: var(--muted);
	}
	tr:nth-child(odd) { background-color: #2a2a3c; }
	tr:hover { background-color: #3a3a4f; }

	label{
		font-size:15px;
		color:var(--muted);
		display:block;
		margin-bottom:6px
	}


	.side-by-side{display:grid; grid-template-columns: 1fr 1fr; gap:12px}

	.order-type{
	flex:1; display:flex; align-items:center; gap:12px; padding:8px 12px; border-radius:12px; cursor:pointer;
	user-select:none; border:1px solid rgba(255,255,255,0.03); background:transparent; font-size:15px;
	margin-bottom: 0;
	}
	.order-type input{display:none}

	.order-type.buy{background:linear-gradient(90deg, rgba(22,163,74,0.08), transparent); border-color: rgba(16,185,129,0.08)}
	.order-type.sell{background:linear-gradient(90deg, rgba(239,68,68,0.06), transparent); border-color: rgba(239,68,68,0.08)}

	.order-type:has(input:checked) {
	  background-color: rgba(52, 199, 89, 0.28); /* green full */
	  border-color: rgba(52, 199, 89, 0.5);
	  color: white;
	}

	.order-type:has(input:checked).sell {
	  background-color: rgba(255, 59, 48, 0.28); /* red full */
	  border-color: rgba(255, 59, 48, 0.5);
	  color: white;
	}


	.price-row{display:flex; align-items:center; gap:8px}
	.currency{padding:8px 10px; border-radius:8px; background:rgba(255,255,255,0.02); font-size:14px; color:var(--muted); border:1px solid rgba(255,255,255,0.02)}

	.meta{display:flex; justify-content:space-between; align-items:center; font-size:14px; color:var(--muted)}

	.total{font-weight:600; font-size:16px}

	select,input[type="number"]{
	width:100%; padding:12px 14px; background:var(--glass); border:1px solid rgba(255,255,255,0.04); color:inherit;
	border-radius:12px; font-size:15px; outline:none;
	}
	select:focus,input[type="number"]:focus{box-shadow:0 4px 18px rgba(124,58,237,0.08); border-color:rgba(124,58,237,0.28)}

	button.submit{
	width:100%; max-width:400px; margin:0 auto; display:block;
	padding:14px 16px; border-radius:12px; border:1px solid rgba(255,255,255,0.04); font-weight:600; font-size:15px; cursor:pointer;
	background:var(--glass); color:inherit; text-align:center;
	box-shadow: inset 0 0 0 0 transparent;
	}

	button.submit.buy{background:linear-gradient(90deg, rgba(124,58,237,0.08), transparent); border-color: rgba(124,58,237,0.28); color:#7C3AED}
	button:disabled{opacity:0.45; cursor:not-allowed}

	button.wipe{
	width:100%; max-width:150px; margin:0 auto; display:block;
	padding:8px 12px; border-radius:12px; border:1px solid rgba(255,255,255,0.04); font-weight:600; font-size:10px; cursor:pointer;
	background:var(--glass); color:inherit; text-align:center;
	box-shadow: inset 0 0 0 0 transparent;
	}

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

	<div class="side-by-side">
		<div>
			<h1>Limit Order Book</h1>
		</div>

		<div>
			<button type="button" id="wipe" class="wipe">Wipe OrderBook</button>
		</div>
	</div>

	<h2>Place Order</h2>
	<div class="container">
		<div class="order-form">
		    <form id="orderForm" novalidate>
		      <div>
			<label for="side">Side</label>
			<div class="row" role="radiogroup" aria-label="Order side">
			  <label class="order-type buy" id="buyOption">
			    <input type="radio" name="side" value="buy" id="sideBuy" checked aria-checked="true">
			    <span aria-hidden>▲</span>
			    <span style="min-width:40px;">Buy</span>
			  </label>

			  <label class="order-type sell" id="sellOption">
			    <input type="radio" name="side" value="sell" id="sideSell" aria-checked="false">
			    <span aria-hidden>▼</span>
			    <span style="min-width:40px;">Sell</span>
			  </label>
			</div>
		      </div>

		      <div class="side-by-side">
			<div>
			  <label for="price">Price</label>
			  <div class="price-row">
			    <input id="price" name="price" type="number" inputmode="decimal" step="1" min="0" placeholder="0" aria-describedby="priceHelp" required>
			  </div>
			</div>

			<div>
			  <label for="size">Size</label>
			  <input id="size" name="size" type="number" inputmode="decimal" step="1" min="0" placeholder="0" aria-describedby="sizeHelp" required>
			</div>
		      </div>

		      <div>
			<button type="submit" id="submit" class="submit buy order-type buy">Place Order</button>
		      </div>
		    </form>
		</div>

		<div class="order-book">
			<div class="side-by-side">
				<div>
					<label>Bids</label>
					<table>
						<tr>
							<th>Price</th>
							<th>Volume</th>
						</tr>
						{{range .Bids}}
						<tr>
							<td>{{.Price}}</td>
							<td>{{.Volume}}</td>
						</tr>
						{{end}}
					</table>
				</div>

				<div>
					<label>Asks</label>
					<table>
						<tr>
							<th>Price</th>
							<th>Volume</th>
						</tr>
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
	</div>

	<div class="order-book">
		<div>
		<label>Trades</label>
		<table>
			<tr>
				<th>Price</th>
				<th>Size</th><th>Time</th>
				<th>Buyer ID</th>
				<th>Seller ID</th>
			</tr>
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
	</div>

	<footer>
		<div class="hint">New Connected Pod: {{.Hostname}}</div>
	</footer>

	<script>

		const radios = document.querySelectorAll('input[name="side"]');
		radios.forEach(radio => {
			radio.addEventListener('change', () => {
				radios.forEach(r => r.setAttribute('aria-checked', r.checked));
				console.log("Selected side:", document.querySelector('input[name="side"]:checked').value);
			});
		});

		const form = document.getElementById('orderForm');
		form.addEventListener('submit', async (e) => {
			e.preventDefault();

			const side = document.querySelector('input[name="side"]:checked').value;
			const price = parseInt(document.getElementById('price').value);
			const size = parseInt(document.getElementById('size').value);

			const resp = await fetch('/api/order', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify({side, price, size})
			});
			if (resp.ok) {
				window.location.reload();
			} else {
				const text = await resp.text();
				alert('Error: ' + text);
			}
		});
		const wipe = document.getElementById('wipe')
		wipe.addEventListener('click', async () => {
		  if (!confirm("Are you sure you want to wipe all orders?")) return;

		  const resp = await fetch('/api/wipe', {
		    method: 'POST'
		  });

		  if (resp.ok) {
		    window.location.reload();
		  } else {
		    const text = await resp.text();
		    alert('Error: ' + text);
		  }
		});
	</script>
	</div>
</body>
</html>`
