package web

const css = `.container {
  width:80vw;
  margin:auto;
  display:block;
}

canvas {
  width:100%;
  height:auto;
}`

const chartTemplate = `var ctx = document.getElementById('chart').getContext('2d');

const dataJSON = ` + "`{{.}}`;" + `
const data = JSON.parse(dataJSON)

var chart = new Chart(ctx, {
    type: 'line',
    data: {
        datasets: [{
            label: 'People',
            yAxisID: 'people',
            data: data.People,
            backgroundColor: '#36a8e1',
            borderColor: 'darkblue'
        },
        {
            label: 'Capacity',
            yAxisID: 'people',
            data: data.Capacity,
            backgroundColor: 'red',
            borderColor: 'red',
            fill: false
        },
        {
            label: 'Percent',
            yAxisID: 'percent',
            data: data.Percent,
			hidden: true,
            backgroundColor: 'black',
            borderColor: 'black',
            fill: false
        }]
    },
    options: {
        padding: 10,
        title: {
            text: "Sputnik",
            display: true,
            fontColor: '#36a8e1',
            fontSize: 20
        },
        elements: {
            line: { tension: 0 },
        },
        scales: {
            xAxes: [{
                type: 'time',
                time: {
                    unit: 'day',
                    displayFormats: {
                        day: 'dddd, MMM D'
                    }
                }
            }],
            yAxes: [{
                id: 'people',
                position: 'left',
                ticks: {
                    min: 0,
                    max: 250
                },
                scaleLabel: {
                    display: true,
                    labelString: '# of people'
                }
            },
            {
                id: 'percent',
                position: 'right',
                ticks: {
                    min: 0,
                    max: 125
                },
                scaleLabel: {
                    display: true,
                    labelString: 'Percent of capacity'
                },
				gridLines: {
                    drawOnChartArea: false // only want the grid lines for one axis to show up
                }
            }]
        },
        tooltips: {
            callbacks: {
                title: function (tooltipItem, data) {
                    var raw = tooltipItem[0].xLabel;
                    var date = new Date(raw);

                    var formatter = new Intl.DateTimeFormat('en-us', {
                        weekday: 'long',
                        month: 'short',
                        day: 'numeric',
                        hour: 'numeric',
                        minute: 'numeric',
                        hour12: false
                    });

                    var result = formatter.format(date);

                    return result;
                },
            }
        }
    }
});`

const popularity = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <title>Sputnik Popularity</title>
  <link rel="stylesheet" href="./style.css">
</head>

<body>

  <div class="container">
    <canvas id="chart"></canvas>
  </div>

</body>

<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.9.3/Chart.bundle.js"></script>
<script src="./chart.js"></script>

</html>`
