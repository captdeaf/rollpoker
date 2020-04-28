function View(attrs) {
  for (var name in attrs) {
    this[name] = attrs[name];
  }
}

View.prototype.init = function() {
  if (this.Templates) {
    this.T = {};
    for (var tplname in this.Templates) {
      this.T[tplname] = _.template($(this.Templates[tplname]).html());
    }
  }
}

View.prototype._handleEvent = function(name, evt) {
  // name == "Click", "Submit", etc. UpperCamelCase.
  var targ = evt.target;
  var handlers = this["On" + name];
  if (!handlers) { return; }
  while (targ != null && targ != undefined) {
    if (handlers[targ.id]) {
      evt.target = targ;
      if (handlers[targ.id].call(this, evt) != true) {
        evt.preventDefault();
        evt.stopPropagation();
        return
      }
    }
    targ = targ.parentElement;
  }
}
