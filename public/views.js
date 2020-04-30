// VIEWS contains the renderers
var VIEWS = {};

function View(attrs) {
  for (var name in attrs) {
    this[name] = attrs[name];
  }
};

View.prototype.init = function() {
  this.Cache = {};
  if (this.Templates) {
    this.T = {};
    for (var tplname in this.Templates) {
      this.T[tplname] = _.template($(this.Templates[tplname]).html());
    }
  }
};

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
};

View.prototype._gameEvent = function(name, args) {
  if (this.OnEvent[name]) {
    this.OnEvent[name].apply(this, args);
  } else {
    console.log("No Events[" + name + "]!");
  }
};

View.prototype.ValueDiffers = function(name, val) {
  var ret = this.Differs(this.Cache[name], val);
  this.Cache[name] = val;
  return ret;
};

View.prototype.Differs = function(alpha, beta) {
  if (typeof alpha == "object") {
    for (var a in alpha) {
      if (this.Differs(alpha[a], beta[a])) return true
    }
    for (var b in beta) {
      if (this.Differs(alpha[b], beta[b])) return true
    }
    return false
  } else {
    return alpha != beta
  }
};
