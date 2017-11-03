new Vue({
    el: '#app',

    data() {
      return {
          ws: null, // Our websocket
          newMsg: '', // Holds new messages to be sent to the server
          chatContent: [], // A running list of chat messages displayed on the screen
          email: null, // Email address used for grabbing an avatar
          username: null, // Our username
          joined: false, // True if email and username have been filled in
          broadcastTitle: '',
          broadcastMsg: '',
          chatEnabled: false,
          disableChat: false,
          disableBroadcast: false
        }
    },

    created: function() {
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);

            if (msg.hasOwnProperty('username')) {
              self.chatContent.push('<div class="chip">'
                      + '<img src="' + self.gravatarURL(msg.email) + '">' // Avatar
                      + msg.username
                      + '</div>'
                      + emojione.toImage(msg.message) + '<br/>'); // Parse emojis
              if (self.chatContent.length > 8) {
                self.chatContent.splice(0, 1);
              }
            } else if (msg.hasOwnProperty('title')) {
              self.broadcastTitle = msg.title
              self.broadcastMsg = msg.message
            }
            if (chatEnabled) {
              var element = document.getElementById('chat-messages');
              element.scrollTop = element.scrollHeight - element.clientHeight + 20000; // Auto scroll to the bottom
            }
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        email: this.email,
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                this.newMsg = ''; // Reset newMsg
            }
        },

        join: function () {
            var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
            if (!this.email || !re.test(this.email)) {
                Materialize.toast('Du må legge til en gyldig email addresse', 2000);
                return
            }
            if (!this.username) {
                Materialize.toast('Du må velge et brukernavn', 2000);
                return
            }
            this.email = $('<p>').html(this.email).text();
            this.username = $('<p>').html(this.username).text();
            this.joined = true;
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});
