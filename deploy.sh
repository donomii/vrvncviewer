echo Removing package from android...
adb uninstall com.praeceptamachinae.vrvncviewer
adb uninstall vrvncviewer
rm vrvncviewer.apk
rm vrvncviewerDeploy/lib/armeabi-v7a/libvrvncviewer.so
echo Building...
./gomobile build github.com/donomii/vrvncviewer
echo (pwd)
rm -rf temp
echo (pwd)
mkdir temp
cd temp
unzip ../vrvncviewer.apk
cd ..
cp temp/lib/armeabi-v7a/libvrvncviewerDebug.so  vrvncviewerDeploy/lib/armeabi-v7a/libvrvncviewer.so
cd vrvncviewerDeploy
ant clean
ant release
adb install bin/vrvncviewer-release.apk
cp bin/vrvncviewer-release.apk ../
#adb install bin/vrvncviewer-debug.apk
