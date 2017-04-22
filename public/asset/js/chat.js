$(function () {
  var wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/love';
  var ws = new WebSocket(wsUrl);
  var typingTimer;

  $('form').submit(function () {
    ws.send(JSON.stringify({
      username: 'Equim',
      type: 'chat',
      message: $('#m').val()
    }));
    $('#m').val('');
    return false;
  });

  function stoppedTyping() {
    ws.send(JSON.stringify({
      username: 'Equim',
      type: 'typing',
      message: 'false'
    }));
  }

  $('#m').on('input', function(e) {
    clearTimeout(typingTimer);

    typingTimer = setTimeout(stoppedTyping, 2000);

    ws.send(JSON.stringify({
      username: 'Equim',
      type: 'typing',
      message: 'true'
    }));
  });

  ws.onmessage = function (e) {
    var data = JSON.parse(e.data);
    switch (data.type) {
      case 'chat':
        var message = util.escapeHtml(data.message);
        $('#messages').append($('<li>').html(emojione.toImage(message)));
        break;
      case 'online users':
        console.log('online:', data.message);
        break;
    }
  };

  ws.onclose = function () {
    alert('DISCONNECTED!');
    // TODO: 重连机制
  };
});
