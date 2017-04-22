$(function () {
  $('form').submit(function (e) {
    e.preventDefault();

    $.ajax({
      type: 'POST',
      url: '/register',
      contentType: 'application/json; charset=utf-8',
      data: JSON.stringify({
        likes: ['yuri', 'loli', 'schoolgirl', 'vanilla']
      }),
      success: function (data) {
        document.location = '/chat';
      }
    });
  });
});
