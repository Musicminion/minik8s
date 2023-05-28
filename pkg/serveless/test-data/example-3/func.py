def main(params):
    x = params["x"]
    y = params["y"]
    xx = x * x
    yy = y * y
    resp = {
        "x": xx,
        "y": yy
    }
    return resp