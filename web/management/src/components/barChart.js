import { Bar } from 'vue-chartjs'

export default {
  extends: Bar,
  data: () => ({
    datacollection: {
      labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
      datasets: [
        {
          label: '',
          backgroundColor: '#0501f0',
          data: [0, 0, 0, 0, 10, 0, 30, 0, 0, 40, 10]
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        xAxes: [{
          scaleLabel: {
            display: true,
            labelString: 'Last 12 months'
          }
        }],
        yAxes: [{
          scaleLabel: {
            display: true,
            labelString: 'Events'
          },
          ticks: {
            beginAtZero: true,
            suggestedMax: 50
          }
        }]
      }
    }
  }),

  mounted () {
    this.renderChart(this.datacollection, this.options)
  }
}
