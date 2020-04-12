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
    Signup.ShowForm();
    Signup.ShowMembers(state);
  },
  ShowMembers: function(state) {
    for (var id in state.Players) {
      var listing = $('#' + id);
      if (listing.length == 0) {
        listing = $("<li>");
        listing.attr('id', id);
        listing.html(Signup.PLAYER(state.Players[id]));
        $('#signuplist').append(listing);
      }
    }
  },
  ShowForm: function() {
    $('#joingamediv').show();
    $('#joingameform').submit(function(evt) {
      var data = {}
      _.each($('#joingameform').serializeArray(), function(fd) {
        data[fd.name] = fd.value;
      });
      // TODO: Sanity check email and display name
      POKER.SendCommand("invite", data);
      evt.preventDefault();
      evt.stopPropagation();
    });
  }
};
