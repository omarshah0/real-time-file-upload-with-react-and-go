import FileUploader from './components/FileUploader'

function App() {
  return (
    <div className='min-h-screen bg-gray-100 py-12 px-4 sm:px-6 lg:px-8'>
      <div className='max-w-md mx-auto'>
        <div className='bg-white shadow sm:rounded-lg p-6'>
          <h1 className='text-2xl font-bold mb-6'>File Upload Demo</h1>
          <FileUploader />
        </div>
      </div>
    </div>
  )
}

export default App
