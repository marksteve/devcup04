import React from 'react'
import superagent from 'superagent'
import cn from 'classnames'
import SoundCloudAudio from 'soundcloud-audio'

var App = React.createClass({
  getInitialState() {
    return {
      channel: null,
      channels: [],
      queue: [],
      playing: null,
    }
  },
  componentDidMount() {
    this.scPlayer = new SoundCloudAudio('392fa845f8cff3705b90006915b15af0');
    superagent.get('/me')
      .end((err, res) => {
        this.setState(res.body, () => {
          this.openChannel(this.state.channels[0])
        })
      })
  },
  componentDidUpdate(prevProps, prevState) {
    if (!prevState.queue.length && this.state.queue.length) {
      let song = this.state.queue[0]
      this.scPlayer.resolve(song.from_url, (track) => {
        this.setState({playing: 0})
        this.scPlayer.play()
      })
    }
    if (prevState.playing !== this.state.playing) {
      let bg = this.state.queue[this.state.playing].thumb_url
      React.findDOMNode(this)
        .querySelector('.queue')
        .style.backgroundImage = `url(${bg})`
    }
  },
  render() {
    return (
      <div className="app">
        <h1>
          <span>R</span>
          <span>a</span>
          <span>d</span>
          <span>i</span>
          <span>o</span>
          Slack
        </h1>
        <ul className="channels">
        {this.state.channels.map(this.renderChannel)}
        </ul>
        <ul className="queue">
          {this.state.queue.map(this.renderSong)}
        </ul>
      </div>
    )
  },
  renderChannel(channel) {
    return (
      <li
        key={channel.id}
        onClick={this.openChannel.bind(this, channel)}
        className={cn({
          'active': this.state.channel ? (
            channel.id === this.state.channel.id
          ) : false,
        })}
      >
        #{channel.name}
      </li>
    )
  },
  openChannel(channel) {
    this.setState({
      channel: channel,
      queue: [],
    })
    if (this.socket) {
      this.socket.close()
      delete this.socket
    }
    let wsHost = location.origin.replace('http', 'ws')
    this.socket = new WebSocket(
      `${wsHost}/queue/${this.state.team_id}/${channel.id}`
    )
    this.socket.onmessage = this.updateQueue
  },
  updateQueue(msg) {
    let song = JSON.parse(msg.data)
    this.setState({
      queue: this.state.queue.concat(song),
    })
  },
  renderSong(song, i) {
    switch (song.service_name) {
      case 'SoundCloud':
        return this.renderSoundCloud(song, i)
    }
  },
  renderSoundCloud(song, i) {
    return (
      <li
        key={i}
        className="soundcloud"
      >
        {this.state.playing === i ? (
        <i className="playing icon-controller-play" />
        ) : null}
        <strong>{song.title}</strong>
        <i className="service icon-soundcloud" />
      </li>
    )
  },
})

window.RadioSlack = {
  open(root) {
    React.render(<App />, root)
  },
}
