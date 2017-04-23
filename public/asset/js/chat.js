$(function () {
  var wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/loveStream';
  var ws = new WebSocket(wsUrl);
  var typingTimer;

  ws.onopen = function () {
    // 注册
    ws.send(JSON.stringify({
      username: 'Equim',
      gender: false,
      likes: ['Identity', 'Yuri', 'identity', 'Loli', 'Schoolgirl', 'Vanilla', 'Loli', 'shit'],   // 大小写敏感
      timezone: 8
      //token:
    }));

    $('form').submit(function () {
      ws.send(JSON.stringify({
        type: 'chat',
        message: $('#m').val()
      }));
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
          $('#messages').append($('<li>').html(emojione.toImage(message)));
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
      // TODO: 重连机制
    };
  };
});
