$(function () {
  $('form').submit(function (e) {
    e.preventDefault();

    $.ajax({
      type: 'POST',
      url: '/access',
      contentType: 'application/json; charset=utf-8',
      /*
      data: JSON.stringify({
        likes: $('#likes').val()
      }),
      */
      data: JSON.stringify({
        likes: ['yuri', 'loli', 'schoolgirl', 'vanilla']
      }),
      success: function (data) {
        document.location = '/chat';
      }
    });
  });
});
