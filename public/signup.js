// Initialize sprites/etc:
var Signup = {
  Templates: {
    VIEW: "#signupview",
    LISTING: "#signuplist",
  },
  Setup: function() {
    for (var key in this.Templates) {
      this[key] = _.template($(this.Templates[key]).html());
    }
  },

  Start: function(state) {
    $('body').html(Signup.VIEW());
    Signup.ShowForm();
  },
  Update: function(state) {
    Signup.ShowMembers(state);
  },
  ShowMembers: function(state) {
    $('#signups').html(Signup.LISTING(state));
  },
  ShowForm: function() {
    $('#startgame').click(function(evt) {
      Poker.SendCommand("StartGame", {});
      evt.preventDefault();
      evt.stopPropagation();
    });
    $('#joingameform').submit(function(evt) {
      var data = {}
      _.each($('#joingameform').serializeArray(), function(fd) {
        data[fd.name] = fd.value;
      });
      // TODO: Sanity check email and display name
      Poker.SendCommand("invite", data);
      evt.preventDefault();
      evt.stopPropagation();
    });
  }
};
