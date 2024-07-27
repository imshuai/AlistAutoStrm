import os

# edit strm file and only save first line of it
def save_first_line(file_path):
    with open(file_path, 'r') as file:
        first_line = file.readline()
        with open(file_path, 'w') as new_file:
            new_file.write(first_line)

def process_strm_file(file_path):
    # Open the file and read its contents, replace the 'https://pan.510222.xyz' to 'http://zuk.v2ns.eu.org:5244', then write back to the file
    with open(file_path, 'r+') as file:
        file_content = file.read()
        file_content = file_content.replace('https://pan.510222.xyz', 'http://zuk.v2ns.eu.org:5244')
        file.seek(0)
        file.write(file_content)
        file.truncate()

def walk_directory(directory):
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.strm'):
                file_path = os.path.join(root, file)
                process_strm_file(file_path)
                print(f"Processed file: {file_path}")

directory = '/media/sda1'  # Replace with the path to your directory
walk_directory(directory)