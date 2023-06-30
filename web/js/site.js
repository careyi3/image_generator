let coords = [];

function drawAt(point) {
  var dotSize = 10; // in px
  var div = document.createElement("div");
  div.style.backgroundColor = "#000";
  div.style.width = dotSize + "px";
  div.style.height = dotSize + "px";
  div.style.position = "absolute";
  div.style.left = point.x - dotSize / 2 + "px";
  div.style.top = point.y - dotSize / 2 + "px";
  div.style.borderRadius = "50%";
  document.getElementById("canvas").appendChild(div);
}

function drawRedAt(point) {
  var dotSize = 10; // in px
  var div = document.createElement("div");
  div.style.backgroundColor = "#F00";
  div.style.width = dotSize + "px";
  div.style.height = dotSize + "px";
  div.style.position = "absolute";
  div.style.left = point.X - dotSize / 2 + "px";
  div.style.top = point.Y - dotSize / 2 + "px";
  div.style.borderRadius = "50%";
  document.getElementById("canvas").appendChild(div);
}

document.getElementById("canvas").onclick = function (e) {
  if (e.target.className != "canvas") {
    return;
  }
  var rect = e.target.getBoundingClientRect();
  var x = e.clientX - rect.left;
  var y = e.clientY - rect.top;
  var point = { x: Math.round(x), y: Math.round(y) };
  coords.push(point);
  drawAt(point);
  console.log("x : " + point.x + " ; y : " + point.y);
};

document.getElementById("clear").onclick = function (_) {
  coords = [];
  document.getElementById("error").innerText = "";
  document.getElementById("formula").innerText = "";
  document.getElementById("canvas").innerHTML = "";
};

document.getElementById("submit").onclick = function (_) {
  var degree = document.getElementById("degree").value;
  var rate = document.getElementById("rate").value;
  var itrs = document.getElementById("itrs").value;
  $.post(
    `/submit?degree=${degree}&rate=${rate}&itrs=${itrs}`,
    JSON.stringify(coords),
    function (data, _) {
      response = JSON.parse(data);
      document.getElementById("error").innerText =
        "Percentage Error: " + response.TrainingError;
      document.getElementById("formula").innerText =
        "Formula: " + response.Formula;
      response.Points.forEach((point) => {
        if (point.X >= 0 && point.X <= 600 && point.Y >= 0 && point.Y <= 600) {
          drawRedAt(point);
        }
      });
    }
  ).fail(function () {
    alert("error");
  });
};
