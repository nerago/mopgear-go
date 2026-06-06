
# Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# winget install Kitware.CMake --source winget
# winget install MSYS2.MSYS2
# C:\msys64\msys2_shell.cmd -defterm -here -no-start -ucrt64
# pacman -S --needed base-devel mingw-w64-ucrt-x86_64-toolchain

# for VSCode setup config
#"go.toolsEnvVars": {
  #"CGO_ENABLED": "1",
  #"CC": "C:/msys64/ucrt64/bin/gcc.exe",
  #"CXX": "C:/msys64/ucrt64/bin/g++.exe",
  #"PATH": "${env:PATH};C:/msys64/ucrt64/bin/" 
#},

# use cmake to build the HiGHS windows .dll
try {
	$Env:Path += ";C:\msys64\ucrt64\bin"

	Remove-Item -Recurse -Force build_windows
	mkdir build_windows
	cd build_windows

	# not possible, CUDA compiler needs MSVC, go needs ming
	# need to switch away from static library at minimum	
	# cmake -G "MinGW Makefiles" -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_BUILD_TYPE=Release -DCMAKE_POSITION_INDEPENDENT_CODE=ON -DBUILD_TESTING=OFF -DCUPDLP_FIND_CUDA=ON -DCUDAToolkit_ROOT="C:/Program Files/NVIDIA GPU Computing Toolkit/CUDA/v13.3" -DCMAKE_CUDA_PATH="C:/Program Files/NVIDIA GPU Computing Toolkit/CUDA/v13.3" -DCUPDLP_GPU=ON ../..

	cmake -G "MinGW Makefiles" -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_BUILD_TYPE=Release -DCMAKE_POSITION_INDEPENDENT_CODE=ON -DBUILD_TESTING=OFF ../..

	cmake --build . --config Release

	rm ..\..\..\gohighs\internal\highs\lib\windows_amd64\*
	copy .\bin\libhighs.a  ..\..\..\gohighs\internal\highs\lib\windows_amd64

	cd ..
} catch {
	Write-Error "Failed to build for Windows: $($_.Exception.Message)"
}
