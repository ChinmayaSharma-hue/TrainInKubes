import tensorflow as tf
from tensorflow.keras import layers, models
import os

# Defining the model architecture, I later have to experiment with transfer learning
model = models.Sequential()
model.add(layers.Conv2D(32, (3, 3), activation='relu', input_shape=(32, 32, 3)))
model.add(layers.MaxPooling2D((2, 2)))
model.add(layers.Conv2D(64, (3, 3), activation='relu'))
model.add(layers.MaxPooling2D((2, 2)))
model.add(layers.Conv2D(64, (3, 3), activation='relu'))
model.add(layers.Flatten())
model.add(layers.Dense(64, activation='relu'))
model.add(layers.Dense(10, activation='softmax'))

# Compiling the model
model.compile(optimizer='adam', loss='sparse_categorical_crossentropy', metrics=['accuracy'])

# Saving the model onto a persistent volume
# Take the location to store the model from the environment variable, have to fix this later
# Hint - Use Environment Variables

# Get the environmental variable MODEL_STORAGE_LOCATION
model_storage_location = os.environ['MODEL_STORAGE_LOCATION']

model.save(model_storage_location + '/model.h5')