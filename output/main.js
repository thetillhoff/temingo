window.onload = function() {

  /* BUTTONS */
  var buttonsByTag = document.getElementsByTagName("button");
  var buttonsByClass = document.getElementsByClassName("button");
  var buttons = [ ...buttonsByTag, ...buttonsByClass];
  for (var i = 0; i < buttons.length; i++) {
    console.log("found a button!")
  }
  
}