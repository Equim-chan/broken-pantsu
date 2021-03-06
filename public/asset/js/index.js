if (!window.WebSocket) {
  alert('Sorry, your browser is too old to run Broken Pantsu!\nPlease use a morden browser that supports WebSocket.');
  location.href = 'http://outdatedbrowser.com';
}

var util = (function(){
  var _util = {};

  var entityMap = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
    '/': '&#x2F;',
    '`': '&#x60;',
    '=': '&#x3D;'
  };

  _util.escapeHtml = function (string) {
    return String(string).replace(/[&<>"'`=\/]/g, function (s) {
      return entityMap[s];
    });
  };

  _util.readCookie = function (name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for(var i = 0; i < ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0) == ' ') c = c.substring(1, c.length);
      if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
  };

  return _util;
})();

// THIS PART IS FOR TEST ONLY
var ws;
function sw() {
  ws.send(JSON.stringify({
    type: 'switch',
    msg: ''
  }));
}
// //////////////////////////

$(function () {
  var token = util.readCookie('token');
  if (!token) {
    location.reload(true);
    return;
  }
  var wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/loveStream';
  ws = new WebSocket(wsUrl);
  var typingTimer;

  ws.onopen = function () {
    // register
    ws.send(JSON.stringify({
      username: 'Equim',
      gender: false,
      likes: ['Identity', 'Yuri', 'identity', 'Loli', 'Schoolgirl', 'Vanilla', 'Loli', 'shit'],   // 大小写敏感
      timezone: 8,
      token: token
    }));

    $('form').submit(function () {
      var message = $('#m').val()
      ws.send(JSON.stringify({
        type: 'chat',
        msg: message
      }));
      message = util.escapeHtml(message);
      $('#messages').append($('<li class="self">').html(emojione.toImage(message)));
      $('#m').val('');
      return false;
    });

    function stoppedTyping() {
      ws.send(JSON.stringify({
        type: 'typing',
        msg: 'false'
      }));
    }

    $('#m').on('input', function(e) {
      clearTimeout(typingTimer);

      typingTimer = setTimeout(stoppedTyping, 2000);

      ws.send(JSON.stringify({
        type: 'typing',
        msg: 'true'
      }));
    });

    ws.onmessage = function (e) {
      console.log(e.data);
      var data = JSON.parse(e.data);
      switch (data.type) {
        case 'chat':
          var message = util.escapeHtml(data.msg);
          $('#messages').append($('<li class="partner">').html(emojione.toImage(message)));
          break;
        case 'online users':
          console.log('online:', data.msg);
          break;
        case 'matched':
          break;
        case 'reject':
          alert(data.msg);
          break;
      }
    };

    ws.onclose = function () {
      alert('DISCONNECTED!');
      // TODO: auto-reconnection, but depends on the circumstance. If the disconnection is caused initiatively
      // by the server, then it must be your own reason and there's no need to retry again :)
    };
  };
});
