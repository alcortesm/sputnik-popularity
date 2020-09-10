package web

const css = `h1 {
  color: #36a8e1;
  text-align: center;
}

.container {
  width:70vw;
  margin:auto;
  display:block;
}

canvas {
  width:100%;
  height:auto;
}`

const chartTemplate = `var ctx = document.getElementById('chart').getContext('2d');

const data = {{.People}};

var chart = new Chart(ctx, {
  type: 'line',
  data: {
    datasets: [{
      label: 'People',
      data: data
    }]
  },
  options: {
    scales: {
      xAxes: [{
        type: 'time'
      }]
    }
  }
});
`

const popularity = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <title>title</title>
  <link rel="stylesheet" href="./style.css">
</head>

<body>

  <h1>Sputnik Popularity</h1>

  <div class="container">
    <canvas id="chart"></canvas>
  </div>

</body>

<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.9.3/Chart.bundle.js"></script>
<script src="./chart.js"></script>

</html>`

const noData = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Sputnik Popularity</h1>
	<p>There are no data available.</p>

</body>
</html>`
