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

$(function () {
  $.ajax({
    type: 'POST',
    url: '/',
    contentType: 'application/json; charset=utf-8',
    data: encodeURI(JSON.stringify({
      username: 'Equim',
      gender: false,
      likes: ['Identity', 'Yuri', 'identity', 'Loli', 'Schoolgirl', 'Vanilla', 'Loli', 'shit'],   // 大小写敏感
      timezone: 8,
      token: token
    }))
  });

  var token = util.readCookie('token');
  if (!token) {
    location.reload(true);
    return;
  }
  var wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/loveStream';
  var ws = new WebSocket(wsUrl);
  var typingTimer;

  ws.onopen = function () {
    // 注册
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
        message: message
      }));
      message = util.escapeHtml(message);
      $('#messages').append($('<li class="self">').html(emojione.toImage(message)));
      $('#m').val('');
      return false;
    });

    function stoppedTyping() {
      ws.send(JSON.stringify({
        type: 'typing',
        message: 'false'
      }));
    }

    $('#m').on('input', function(e) {
      clearTimeout(typingTimer);

      typingTimer = setTimeout(stoppedTyping, 2000);

      ws.send(JSON.stringify({
        type: 'typing',
        message: 'true'
      }));
    });

    ws.onmessage = function (e) {
      console.log(e.data);
      var data = JSON.parse(e.data);
      switch (data.type) {
        case 'chat':
          var message = util.escapeHtml(data.message);
          $('#messages').append($('<li class="partner">').html(emojione.toImage(message)));
          break;
        case 'online users':
          console.log('online:', data.message);
          break;
        case 'matched':
          console.log(data.partnerInfo);
          break;
        case 'reject':
          alert(data.message);
          break;
      }
    };

    ws.onclose = function () {
      alert('DISCONNECTED!');
      // TODO: 重连机制，但是要看情况！如果是服务器主动把连接断了怎么想都是你自己的原因嘛所以这种情况就不要重试了。
    };
  };
});
