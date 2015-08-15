import React from 'react'

var App = React.createClass({
  render() {
    return (
      <div className="app">
        <h1>RadioSlack</h1>
      </div>
    )
  },
})

window.RadioSlack = {
  open(root) {
    React.render(<App />, root)
  },
}
