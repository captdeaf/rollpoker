var Helpers = {
  RaiseImage: function(offset, cls) {
    var startoff = {
      left: offset.left,
      top: offset.top + (40 * RollPoker.Scale),
    }
    var adiv = $("<div>");
    adiv.addClass("eventanim");
    adiv.addClass("raiseimage");
    if (cls) {
      adiv.addClass(cls);
    }
    adiv.offset(startoff);
    var newoffset = {
      left: offset.left,
      top: offset.top - (20 * RollPoker.Scale),
    }
    adiv.css({
      "transform-origin": "0 0",
      "-webkit-transform": RollPoker.Scales,
      "-moz-transform": RollPoker.Scales,
      "-ms-transform": RollPoker.Scales,
      "-o-transform": RollPoker.Scales,
      "transform": RollPoker.Scales,
    });
    $("body").append(adiv);
    adiv.animate(newoffset, 1200, "swing", function() {
      adiv.remove();
    });
  },
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
    $("body").append(adiv);
    adiv.animate(newoffset, 1200, "swing", function() {
      adiv.remove();
    });
  },
};
