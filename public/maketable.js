$(document).ready(function() {
  $("#creategame").click(function() {
    var data = {}
    _.each($('#gamesettings').serializeArray(), function(fd) {
      data[fd.name] = fd.value;
    });
    console.log(data);
    $.ajax({
      url: '/MakeTable',
      type: 'POST',
      dataType: 'json',
      data: JSON.stringify(data),
      success: function(result) {
        console.log(result);
        document.location = "/table/" + result.Name;
      },
      error: function(xhr, resp, text) {
        console.log(xhr, resp, text);
      },
    })
  });
});
