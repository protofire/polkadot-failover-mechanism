locals {

  lb_hc = {
    combo = ["Http", "8080", "/verify/checks"]
  }

  lb_port = {
    http     = ["8500", "Tcp", "8500"]
    dns      = ["8600", "Tcp", "8600"]
    rpc      = ["8300", "Tcp", "8300"]
    lan      = ["8301", "Tcp", "8301"]
    wan      = ["8302", "Tcp", "8302"]
    polkadot = ["30333", "Tcp", "30333"]
  }

}
