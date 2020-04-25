var Helpers = {
  RaiseText: function(offset, text, cls) {
    var startoff = {
      left: offset.left,
      top: offset.top + 40,
    }
    var adiv = $("<div>");
    adiv.text(text);
    adiv.addClass("eventanim");
    if (cls) {
      adiv.addClass(cls);
    }
    adiv.offset(startoff);
    var newoffset = {
      left: offset.left,
      top: offset.top - 20,
    }
    $("#tables").append(adiv);
    adiv.animate(newoffset, 1200, "swing", function() {
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
      if (amt == 0) {
        Helpers.RaiseText(off, opt);
      } else {
        Helpers.RaiseText(off, opt + " (" + amt + ")");
      }
    }
  },
  Bet: function(playerid, amt, opt) {
    var off = Poker.GetPlayerLocation(playerid);
    Table.MaybeClearQueue();
    if (off) {
      Helpers.RaiseText(off, opt + " " + amt);
    }
  },
  Win: function(playerid, amt, handname, handcards) {
    var off = Poker.GetPlayerLocation(playerid);
    Table.QueueCancel();
    if (off) {
      Helpers.RaiseText(off, "Won " + amt);
    }
  }
};
