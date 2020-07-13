import argparse

def parse_xml_file(filename):
    xml_data = None
    try:
        with open(filename, 'r') as xml_file:
            xml_data = xml_file.read()
    except FileNotFoundError:
        return None

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--filename", type=str)
    args = parser.parse_args()
    filename = args.filename
    parse_xml_file(filename)