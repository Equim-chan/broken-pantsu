$(function () {
  var wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/love';
  var ws = new WebSocket(wsUrl);

  $('form').submit(function () {
    ws.send(JSON.stringify({
      username: 'Equim',
      message: $('#m').val()
    }));
    $('#m').val('');
    return false;
  });

  ws.onmessage = function (e) {
    var data = JSON.parse(e.data);
    var message = util.escapeHtml(data.message);
    $('#messages').append($('<li>').html(emojione.toImage(message)));
  };

  ws.onclose = function () {
    alert('DISCONNECTED!');
    // TODO: 重连机制
  };
});
