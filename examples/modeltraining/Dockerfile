FROM tensorflow/tensorflow:latest-gpu-py3

# Install numpy
RUN pip install numpy

# Copy the training script
COPY ./trainingthemodel.py /script.py

# Run the training script
CMD ["python3", "/script.py"]