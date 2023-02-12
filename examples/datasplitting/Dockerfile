FROM python:3.7-slim

# Install dependencies
RUN pip install numpy

# Copy the script
COPY ./getting_preprocessed_data.py /script.py

# Run the script
CMD ["python3", "/script.py"]