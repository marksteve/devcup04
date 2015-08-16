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
    this.scPlayer.on('ended', this.nextSong)
    superagent.get('/me')
      .end((err, res) => {
        this.setState(res.body, () => {
          this.openChannel(this.state.channels[0])
        })
      })
  },
  componentDidUpdate(prevProps, prevState) {
    // Start playing
    if (!prevState.queue.length && this.state.queue.length) {
      this.setState({playing: 1})
    }
    let {playing} = this.state
    if (this.state.playing && prevState.playing !== playing) {
      let song = this.state.queue[playing - 1]
      if (!song) {
        return
      }
      this.scPlayer.resolve(song.from_url, () => {
        this.scPlayer.play()
      })
      React.findDOMNode(this)
        .querySelector('.queue')
        .style.backgroundImage = `url(${song.thumb_url})`
    }
  },
  render() {
    return (
      <div className="app">
        <header>
          <h1>
            <span>R</span>
            <span>a</span>
            <span>d</span>
            <span>i</span>
            <span>o</span>
            Slack
          </h1>
          <div className="me">
            {this.state.user}
          </div>
        </header>
        <div className="channels">
          <h2>{this.state.team}</h2>
          <ul>
          {this.state.channels.map(this.renderChannel)}
          </ul>
        </div>
        <ul className="queue">
        {this.state.queue.length ? (
          this.state.queue.map(this.renderSong)
        ) : (
          <div className="empty-queue">
            Post a Soundcloud song in this channel
            to get your station started.
          </div>
        )}
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
    this.scPlayer.stop()
    this.setState({
      channel: channel,
      queue: [],
      playing: null,
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
    let playing = this.state.playing - 1 === i
    return (
      <li
        key={i}
        className={cn('soundcloud', {'playing': playing})}
      >
        {playing ? (
        <i className="play icon-controller-play" />
        ) : null}
        {song.title}
        <i className="service icon-soundcloud" />
      </li>
    )
  },
  nextSong() {
    this.setState({playing: this.state.playing + 1})
  },
})

window.RadioSlack = {
  open(root) {
    React.render(<App />, root)
  },
}
