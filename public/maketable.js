$(document).ready(function() {
  $("#creategame").click(function() {
    alert("ok");
    $.ajax({
      url: '/MakeTable',
      type: 'POST',
      dataType: 'json',
      data: $('#gamesettings').serialize(),
      success: function(result) {
        alert(result);
      },
      error: function(xhr, resp, text) {
        console.log(xhr, resp, text);
      },
    })
  });
});
