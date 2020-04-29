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
};
