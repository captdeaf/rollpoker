// Initialize sprites/etc:
var Signup = {
  Templates: {
    LISTING: "#signuptpl",
    PLAYER: "#signupplayertpl",
  },
  Setup: function() {
    for (var key in this.Templates) {
      this[key] = _.template($(this.Templates[key]).html());
    }
  },

  Start: function(state) {
    $('body').html(Signup.LISTING());
    if (POKER.GetPlayerId()) {
      Signup.ShowLoggedIn();
    } else {
      Signup.ShowNotLoggedIn();
    }
  },
  ShowLoggedIn: function() {
    $('.gamebutton').hide();
    $('#leavegame').show();
    $('#startgame').show();
  },
  ShowNotLoggedIn: function() {
    $('.gamebutton').hide();
    $('#joingamediv').show();
    $('#joingame').click(function() {
      var data = {}
      _.each($('#joingameform').serializeArray(), function(fd) {
        data[fd.name] = fd.value;
      });
      // TODO: Sanity check email and display name
      POKER.SendCommand("register", data);
      return false;
    });
  }
};
