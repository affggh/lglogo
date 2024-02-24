with open("raw_resources_a.img", "rb") as fin, \
     open("test.rle", "wb") as fout:
    fin.seek(0x2000, 0)
    fout.write(fin.read(0x903c))
    with open("test.raw", "wb") as fraw:
        fin.seek(0x2000, 0)
        for i in range(0, 0x903c, 4):
            data = fin.read(4)
            for j in range(data[0]):
                fraw.write(data[1:])