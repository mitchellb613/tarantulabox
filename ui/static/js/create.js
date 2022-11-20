var uploadField = document.getElementById("tarantulaImage");

uploadField.onchange = function () {
  if (this.files[0].size > 50 * 1024 * 1024) {
    alert("File is too big!");
    this.value = "";
  }
};
