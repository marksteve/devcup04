import React from 'react'
import superagent from 'superagent'
import cn from 'classnames'
import SoundCloudAudio from 'soundcloud-audio'
import indexBy from 'lodash/collection/indexBy'

var App = React.createClass({
  getInitialState() {
    return {
      showLogin: true,
      channel: null,
      users: [],
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
        if (res.ok) {
          this.setState({showLogin: false})
        }
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
    if (!prevState.users.length && this.state.users.length) {
      this.users = indexBy(this.state.users, 'id')
    }
  },
  render() {
    return (
      <div className="app">
        <header>
          <h1>
            <span>Radio</span>
            <span>S</span>
            <span>l</span>
            <span>a</span>
            <span>c</span>
            <span>k</span>
          </h1>
          {this.state.user ? (
          <div className="me">
            {this.state.user} &mdash; <a href="/logout">Logout</a>
          </div>
          ) : null}
        </header>
        {this.state.showLogin ? (
        <div className="login">
          <p>
            RadioSlack turns your Slack channels into radio stations.<br />
            Just post a song link in the channel to queue songs.<br />
          </p>
          <p>
            <button><a href="/login">Login with Slack</a></button>
          </p>
        </div>
        ) : (
        <div>
          <div className="channels">
            <h2>{this.state.team}</h2>
            <ul>
            {this.state.channels.map(this.renderChannel)}
            </ul>
          </div>
          <ul className="queue">
          {this.state.queue.length ? (
            this.state.queue.map(this.renderSong)
          ) : this.state.channel ? (
            <div className="empty-queue">
              Post a Soundcloud song in this channel
              to get your station started.
            </div>
          ) : null}
          </ul>
        </div>
        )}
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
        <span className="song-title">{song.title}</span>
        <div className="poster">
          <i className="service icon-soundcloud" />
          <span>{this.users[song.user].name}</span>
        </div>
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
