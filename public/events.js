var Helpers = {
  RaiseText: function(offset, text, cls) {
    var adiv = $("<div>");
    adiv.text(text);
    adiv.addClass("eventanim");
    if (cls) {
      adiv.addClass(cls);
    }
    adiv.offset(offset);
    var newoffset = {
      left: offset.left,
      top: offset.top - 20,
    }
    $("#tables").append(adiv);
    adiv.animate(newoffset, 400, "swing", function() {
      adiv.remove();
    });
  },
  RaiseTextFor(playerid, text, cls) {
    var off = Poker.GetPlayerLocation(playerid);
    Helpers.RaiseText(off, text, cls);
  }
};

var Events = {
  Call: function(playerid, amt, opt) {
    var off = Poker.GetPlayerLocation(playerid);
    if (off) {
      Helpers.RaiseText(off, opt + " (" + amt + ")");
    }
  },
  Bet: function(playerid, amt, opt) {
    var off = Poker.GetPlayerLocation(playerid);
    if (off) {
      Helpers.RaiseText(off, opt + " " + amt);
    }
  }
};
