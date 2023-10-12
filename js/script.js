"use strict";

const checkboxSort = document.querySelector("#checkbox-sort");
const checkboxBlack = document.querySelector("#checkbox-black");
const checkboxRed = document.querySelector("#checkbox-red");
const checkboxYellow = document.querySelector("#checkbox-yellow");
const checkboxGreen = document.querySelector("#checkbox-green");

const buttonsOpen = document.querySelectorAll(".button-open");
const buttonsDownload = document.querySelectorAll(".button-download");

function call_backend(action, path) {
  const searchParams = new URLSearchParams();
  if (checkboxSort.checked) {
    searchParams.set("sort", "on");
  }
  if (checkboxBlack.checked) {
    searchParams.set("black", "on");
  }
  if (checkboxRed.checked) {
    searchParams.set("red", "on");
  }
  if (checkboxYellow.checked) {
    searchParams.set("yellow", "on");
  }
  if (checkboxGreen.checked) {
    searchParams.set("green", "on");
  }
  if (action) {
    searchParams.set("action", action);
  }
  if (path) {
    searchParams.set("path", path);
  }
  location.search = searchParams.toString();
}

for (const checkbox of [checkboxSort, checkboxBlack, checkboxRed, checkboxYellow, checkboxGreen]) {
  checkbox.addEventListener("change", () => call_backend());
}

for (const button of buttonsOpen) {
  button.addEventListener("click", () => call_backend("open", button.getAttribute("data-path")));
}

for (const button of buttonsDownload) {
  button.addEventListener("click", () => call_backend("download", button.getAttribute("data-path")));
}
